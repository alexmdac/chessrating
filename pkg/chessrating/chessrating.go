package chessrating

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sajari/regression"
)

// PredictRatingParams are parameters for PredictRating.
type PredictRatingParams struct {
	User         string `json:"user"`           // the chess.com user
	GameType     string `json:"game_type"`      // the chess.com game type.
	DaysAgo      int    `json:"days_ago"`       // how many days of history to use
	DaysInFuture int    `json:"days_in_future"` // how far to look in the future
	Correct      bool   `json:"correct"`        // correct the prediction so that current rating is accurate
}

// DefaultPredictRatingParams are default PredictRatingParams.
var DefaultPredictRatingParams = PredictRatingParams{
	GameType:     "blitz",
	DaysAgo:      30,
	DaysInFuture: 30,
}

// Validate checks that the parameters are valid.
func (p PredictRatingParams) Validate() error {
	if p.User == "" {
		return errors.New("must specify User")
	}
	return nil
}

// PredictRating predicts a user's chess.com rating.
func PredictRating(p PredictRatingParams) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	v := url.Values{}
	v.Add("type", p.GameType)
	v.Add("daysAgo", strconv.Itoa(p.DaysAgo))

	// Obtained by inspecting chess.com's HTML. Quite likely to break!
	url := fmt.Sprintf("https://www.chess.com/callback/live/stats/%s/chart?%s", p.User, v.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("could not get ratings from chess.com: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("GET returned %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("could not read response from chess.com: %v", err)
	}

	var ratings []struct {
		Timestamp, Rating float64
	}
	if err := json.Unmarshal(bodyBytes, &ratings); err != nil {
		return 0, fmt.Errorf("could not unmarshal JSON from chess.com: %v", err)
	}

	if len(ratings) == 0 {
		return 0, fmt.Errorf("user %s has no rating history", p.User)
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

	if p.Correct {
		currentPred, err := reg.Predict([]float64{float64(now.Unix())})
		if err != nil {
			return 0, fmt.Errorf("cannot predict the present: %v", err)
		}
		currentRating := ratings[len(ratings)-1].Rating
		correction = currentRating - currentPred
	}

	futureTime := now.Add(time.Duration(p.DaysInFuture*24) * time.Hour)
	pred, err := reg.Predict([]float64{float64(futureTime.Unix())})
	if err != nil {
		return 0, fmt.Errorf("could not predict the future: %v", err)
	}
	correctedPred := int(pred + correction)

	return correctedPred, nil
}
