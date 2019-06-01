# chess.com rating predictor

This app retrieves a [chess.com](http://chess.com) user's rating history, and
predicts the user's future rating using linear regression.

## Example

To predict user the rating of `alexmdac` in 30 days based on 30 days of history:
```
$ chessrating -user alexmdac -days-ago 30 -days-in-future 30
1180
```

## Rating Correction

It is possible to estimate what a user's rating should be today by specifying
`-days-in-future 0`. This can determine if a user is currently under- or
over-rated. Specify the `-correct` flag to adjust predictions so that the
prediction for today is equal to the user's current rating.