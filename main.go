package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/sajari/regression"
)

var (
	user         = flag.String("user", "", "chess.com username")
	gameType     = flag.String("game-type", "blitz", "type of game")
	daysAgo      = flag.Int("days-ago", 30, "how many days of history to fetch")
	daysInFuture = flag.Int("days-in-future", 90, "how far in the future to predict")
	correct      = flag.Bool("correct", false, "assume that current rating is accurate, and correct estimates")
)

func main() {
	flag.Parse()
	if err := extrapolateRating(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
	}
}

func extrapolateRating() error {
	if *user == "" {
		return errors.New("must specify -user")
	}

	v := url.Values{}
	v.Add("type", *gameType)
	v.Add("daysAgo", strconv.Itoa(*daysAgo))

	// Obtained by inspecting chess.com's HTML. Quite likely to break!
	url := fmt.Sprintf("https://www.chess.com/callback/live/stats/%s/chart?%s", *user, v.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("could not get ratings from chess.com: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GET returned %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response from chess.com: %v", err)
	}

	var ratings []struct {
		Timestamp, Rating float64
	}
	if err := json.Unmarshal(bodyBytes, &ratings); err != nil {
		return fmt.Errorf("could not unmarshal JSON from chess.com: %v", err)
	}

	if len(ratings) == 0 {
		return fmt.Errorf("user %s has no rating history", *user)
	}

	reg := &regression.Regression{}
	reg.SetObserved("Rating")
	reg.SetVar(0, "timestamp")
	for _, rating := range ratings {
		reg.Train(regression.DataPoint(rating.Rating, []float64{rating.Timestamp / 1000}))
	}
	reg.Run()

	now := time.Now()
	correction := 0.0

	if *correct {
		currentPred, err := reg.Predict([]float64{float64(now.Unix())})
		if err != nil {
			return fmt.Errorf("cannot predict the present: %v", err)
		}
		currentRating := ratings[len(ratings)-1].Rating
		correction = currentRating - currentPred
	}

	futureTime := now.Add(time.Duration(*daysInFuture*24) * time.Hour)
	pred, err := reg.Predict([]float64{float64(futureTime.Unix())})
	if err != nil {
		return fmt.Errorf("could not predict the future: %v", err)
	}
	correctedPred := int(pred + correction)
	fmt.Println(correctedPred)

	return nil
}
