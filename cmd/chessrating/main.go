package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alexmdac/chessrating/pkg/chessrating"
)

func main() {
	p := chessrating.DefaultPredictRatingParams

	flag.StringVar(&p.User, "user", p.User, "chess.com username")
	flag.StringVar(&p.GameType, "game-type", p.GameType, "type of game")
	flag.IntVar(&p.DaysAgo, "days-ago", p.DaysAgo, "how many days of history to fetch")
	flag.IntVar(&p.DaysInFuture, "days-in-future", p.DaysInFuture, "how far in the future to predict")
	flag.BoolVar(&p.Correct, "correct", p.Correct, "assume that current rating is accurate, and correct estimates")

	flag.Parse()

	predictedRating, err := chessrating.PredictRating(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	fmt.Printf("%d\n", predictedRating)
}
