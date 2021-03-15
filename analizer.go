package tinvestanalyser

import (
	"github.com/ivangurin/tinvest-client-go"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"time"
)

type Analyser struct {
	Client tinvestclient.Client
}

type Profit struct {
	Ticker         string
	Text           string
	Currency       string
	QuantityBuy    float64
	PriceBuy       float64
	ValueBuy       float64
	CommissionBuy  float64
	QuantitySell   float64
	PriceSell      float64
	ValueSell      float64
	CommissionSell float64
	QuantityEnd    float64
	PriceEnd       float64
	ValueEnd       float64
	DividentValue  float64
	DividentTax    float64
	CouponValue    float64
	CouponTax      float64
	TotalValue     float64
	TotalPercent   float64
	Operations     []tinvestclient.Operation
}

type Signal struct {
	Ticker     string
	Indicators []string
}

func (self *Analyser) Init(ivToken string) {

	self.Client = tinvestclient.Client{}

	self.Client.Init(ivToken)

}

func (self *Analyser) GetProfit(ivTicker string, ivFrom time.Time, ivTo time.Time) (rtProfit []Profit, roError error) {

	ltOperations, roError := self.Client.GetOperations(ivTicker, ivFrom, ivTo)

	if roError != nil {
		return
	}

	ltInstruments := mapOperations(&ltOperations)

	for lvFIGI, ltOperations := range ltInstruments {

		time.Sleep(time.Second)

		lsInstrument, loError := self.Client.GetInstrumentByFIGI(lvFIGI)

		if loError != nil {
			roError = loError
			return
		}

		if lsInstrument.Type == tinvestclient.InstumentTypeCurrency {
			continue
		}

		ltCandles, loError := self.Client.GetCandles(lsInstrument.Ticker, tinvestclient.IntervalDay, ivTo.AddDate(0, 0, -15), ivTo)

		if loError != nil {
			roError = loError
			return
		}

		lsProfit := Profit{}

		lsProfit.Ticker = lsInstrument.Ticker
		lsProfit.Text = lsInstrument.Text
		lsProfit.Currency = lsInstrument.Currency
		lsProfit.Operations = ltOperations

		for _, lsOperation := range ltOperations {

			switch lsOperation.Type {

			case tinvestclient.OperationBuy:

				lsProfit.QuantityBuy += lsOperation.Quantity
				lsProfit.ValueBuy += lsOperation.Value
				lsProfit.CommissionBuy += lsOperation.Commission

			case tinvestclient.OperationBuyCard:

				lsProfit.QuantityBuy += lsOperation.Quantity
				lsProfit.ValueBuy += lsOperation.Value
				lsProfit.CommissionBuy += lsOperation.Commission

			case tinvestclient.OperationSell:

				lsProfit.QuantitySell += lsOperation.Quantity
				lsProfit.ValueSell += lsOperation.Value
				lsProfit.CommissionSell += lsOperation.Commission

			case tinvestclient.OperationDividend:
				lsProfit.DividentValue += lsOperation.Value

			case tinvestclient.OperationTaxDividend:
				lsProfit.DividentTax += lsOperation.Value

			case tinvestclient.OperationCoupon:
				lsProfit.CouponValue += lsOperation.Value

			case tinvestclient.OperationTaxCoupon:
				lsProfit.CouponTax += lsOperation.Value

			}

		}

		if lsProfit.QuantityBuy != 0 {
			lsProfit.PriceBuy = lsProfit.ValueBuy / lsProfit.QuantityBuy
		}

		if lsProfit.QuantitySell != 0 {
			lsProfit.PriceSell = lsProfit.ValueSell / lsProfit.QuantitySell
		}

		lsProfit.QuantityEnd = lsProfit.QuantityBuy - lsProfit.QuantitySell

		if lsProfit.QuantityEnd != 0 {
			lsProfit.PriceEnd = ltCandles[len(ltCandles)-1].Close
			lsProfit.ValueEnd = lsProfit.QuantityEnd * lsProfit.PriceEnd
		}

		lsProfit.TotalValue = lsProfit.ValueEnd -
			lsProfit.QuantityEnd*lsProfit.PriceBuy +
			lsProfit.QuantitySell*lsProfit.PriceSell -
			lsProfit.QuantitySell*lsProfit.PriceBuy -
			lsProfit.CommissionBuy -
			lsProfit.CommissionSell +
			lsProfit.DividentValue - lsProfit.DividentTax +
			lsProfit.CouponValue - lsProfit.CouponTax

		lsProfit.TotalPercent = lsProfit.TotalValue / lsProfit.ValueBuy * 100.

		rtProfit = append(rtProfit, lsProfit)

	}

	return

}

func (self *Analyser) getSignals(itTickers []string) (rtSignals []Signal, roError error) {

	for _, lvTicker := range itTickers {

		time.Sleep(100 * time.Millisecond)

		ltCandles, loError := self.Client.GetCandles(lvTicker, tinvestclient.IntervalDay, time.Now().AddDate(0, 0, -60), time.Now())

		if loError != nil {
			roError = loError
			return
		}

		lsSignal := Signal{}

		lsSignal.Ticker = lvTicker

		if len(ltCandles) == 0 {
			continue
		}

		if isBullishGAP(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "BullishGAP")
		}

		if isBearishGAP(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "BearishGAP")
		}

		if isHammer(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "Hammer")
		}

		if isStar(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "Star")
		}

		if isBullishEngulfing(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "BullishEngulfing")
		}

		if isBearishEngulfing(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "BearishEngulfing")
		}

		if isBullishTweezers(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "BullishTweezers")
		}

		if isBearishTweezers(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "BearishTweezers")
		}

		if isMACDBuy(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "MACD Buy")
		}

		if isMACDSell(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "MACD Sell")
		}

		if isRSIBuy(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "RSI Buy")
		}

		if isRSISell(ltCandles) {
			lsSignal.Indicators = append(lsSignal.Indicators, "RSI Sell")
		}

		if len(lsSignal.Indicators) == 0 {
			continue
		}

		rtSignals = append(rtSignals, lsSignal)

	}

	return

}

func isBullishGAP(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 2 {
		return
	}

	lsPrevCandle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsPrevCandle.High < lsLastCandle.Low {
		rvIs = true
	}

	return

}

func isBearishGAP(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 2 {
		return
	}

	lsPrevCandle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsPrevCandle.Low > lsLastCandle.High {
		rvIs = true
	}

	return

}

func isHammer(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 2 {
		return
	}

	lsPrevCandle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsPrevCandle.Low < lsLastCandle.Open {
		return
	}

	if lsLastCandle.Open == lsLastCandle.Close {
		return
	}

	if lsLastCandle.ShadowHigh > (lsLastCandle.Body / 2) {
		return
	}

	if lsLastCandle.Body == 0 {
		return
	}

	if lsLastCandle.ShadowLow/lsLastCandle.Body < 2 {
		return
	}

	rvIs = true

	return

}

func isStar(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 2 {
		return
	}

	lsPrevCandle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsPrevCandle.High > lsLastCandle.Open {
		return
	}

	if lsLastCandle.Open == lsLastCandle.Close {
		return
	}

	if lsLastCandle.ShadowLow > (lsLastCandle.Body / 2) {
		return
	}

	if lsLastCandle.Body == 0 {
		return
	}

	if (lsLastCandle.ShadowHigh / lsLastCandle.Body) < 2 {
		return
	}

	rvIs = true

	return

}

func isBullishEngulfing(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 2 {
		return
	}

	lsPrevCandle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsLastCandle.Type != tinvestclient.CandleTypeGreen {
		return
	}

	if lsPrevCandle.Type != tinvestclient.CandleTypeRed {
		return
	}

	if lsPrevCandle.Body == 0 {
		return
	}

	if (lsLastCandle.Body / lsPrevCandle.Body) < 3 {
		return
	}

	rvIs = true

	return

}

func isBearishEngulfing(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 2 {
		return
	}

	lsPrevCandle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsLastCandle.Type != tinvestclient.CandleTypeRed {
		return
	}

	if lsPrevCandle.Type != tinvestclient.CandleTypeGreen {
		return
	}

	if lsPrevCandle.Body == 0 {
		return
	}

	if (lsLastCandle.Body / lsPrevCandle.Body) < 3 {
		return
	}

	rvIs = true

	return

}

func isBullishTweezers(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 3 {
		return
	}

	lsPrev2Candle := itCandles[len(itCandles)-3]
	lsPrev1Candle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsLastCandle.Type != tinvestclient.CandleTypeGreen {
		return
	}

	if lsPrev1Candle.Type != tinvestclient.CandleTypeRed &&
		lsPrev2Candle.Type != tinvestclient.CandleTypeRed {
		return
	}

	if lsLastCandle.Low != lsPrev1Candle.Low &&
		lsLastCandle.Low != lsPrev2Candle.Low {
		return
	}

	rvIs = true

	return

}

func isBearishTweezers(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	if len(itCandles) < 3 {
		return
	}

	lsPrev2Candle := itCandles[len(itCandles)-3]
	lsPrev1Candle := itCandles[len(itCandles)-2]
	lsLastCandle := itCandles[len(itCandles)-1]

	if lsLastCandle.Type != tinvestclient.CandleTypeRed {
		return
	}

	if lsPrev1Candle.Type != tinvestclient.CandleTypeGreen &&
		lsPrev2Candle.Type != tinvestclient.CandleTypeGreen {
		return
	}

	if lsLastCandle.High != lsPrev1Candle.High &&
		lsLastCandle.High != lsPrev2Candle.High {
		return
	}

	rvIs = true

	return

}

func isMACDBuy(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	loSeries := techan.NewTimeSeries()

	for _, lsCandle := range itCandles {

		loCandle := techan.NewCandle(techan.TimePeriod{lsCandle.Time, lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loMACDIndincator := techan.NewMACDIndicator(loClosePriceIndicator, 12, 26)

	loMACDHistogram := techan.NewMACDHistogramIndicator(loMACDIndincator, 9)

	lvPrevValue1 := loMACDHistogram.Calculate(len(itCandles) - 1).Float()
	lvPrevValue2 := loMACDHistogram.Calculate(len(itCandles) - 2).Float()
	lvPrevValue3 := loMACDHistogram.Calculate(len(itCandles) - 3).Float()

	if lvPrevValue1 < 0 &&
		lvPrevValue1 > lvPrevValue2 &&
		lvPrevValue3 > lvPrevValue2 {
		rvIs = true
	}

	return

}

func isMACDSell(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	loSeries := techan.NewTimeSeries()

	for _, lsCandle := range itCandles {

		loCandle := techan.NewCandle(techan.TimePeriod{lsCandle.Time, lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loMACDIndincator := techan.NewMACDIndicator(loClosePriceIndicator, 12, 26)

	loMACDHistogram := techan.NewMACDHistogramIndicator(loMACDIndincator, 9)

	lvPrevValue1 := loMACDHistogram.Calculate(len(itCandles) - 1).Float()
	lvPrevValue2 := loMACDHistogram.Calculate(len(itCandles) - 2).Float()
	lvPrevValue3 := loMACDHistogram.Calculate(len(itCandles) - 3).Float()

	if lvPrevValue1 >= 0 &&
		lvPrevValue1 < lvPrevValue2 &&
		lvPrevValue3 < lvPrevValue2 {
		rvIs = true
	}

	return

}

func isRSIBuy(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	loSeries := techan.NewTimeSeries()

	for _, lsCandle := range itCandles {

		loCandle := techan.NewCandle(techan.TimePeriod{lsCandle.Time, lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loRSIIndincator := techan.NewRelativeStrengthIndexIndicator(loClosePriceIndicator, 14)

	if loRSIIndincator.Calculate(len(itCandles)-3).Float() <= 30 {
		rvIs = true
	}

	return

}

func isRSISell(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	loSeries := techan.NewTimeSeries()

	for _, lsCandle := range itCandles {

		loCandle := techan.NewCandle(techan.TimePeriod{lsCandle.Time, lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loRSIIndincator := techan.NewRelativeStrengthIndexIndicator(loClosePriceIndicator, 14)

	if loRSIIndincator.Calculate(len(itCandles)-3).Float() >= 70 {
		rvIs = true
	}

	return

}

func mapOperations(itOperations *[]tinvestclient.Operation) (rtOperations map[string][]tinvestclient.Operation) {

	rtOperations = make(map[string][]tinvestclient.Operation)

	for _, lsOpearion := range *itOperations {

		ltOperations, _ := rtOperations[lsOpearion.FIGI]

		ltOperations = append(ltOperations, lsOpearion)

		rtOperations[lsOpearion.FIGI] = ltOperations

	}

	return

}
