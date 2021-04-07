package tinvestanalyser

import (
	"time"

	tinvestclient "github.com/ivangurin/tinvest-client-go"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type Analyser struct {
	Client tinvestclient.Client
}

type Totals struct {
	PayIn           float64
	PayOut          float64
	ValueBuy        float64
	ValueSell       float64
	ValueTotal      float64
	CommissionBuy   float64
	CommissionSell  float64
	CommissionTotal float64
	DividendValue   float64
	DividendTax     float64
	CouponValue     float64
	CouponTax       float64
	TotalValue      float64
	TotalPercent    float64
}

type Profit struct {
	Ticker         string  `json:"ticker"`
	Text           string  `json:"text"`
	Currency       string  `json:"currency"`
	QuantityBuy    float64 `json:"quantityBuy"`
	PriceBuy       float64 `json:"priceBuy"`
	ValueBuy       float64 `json:"valueBuy"`
	CommissionBuy  float64 `json:"commissionBuy"`
	QuantitySell   float64 `json:"quantitySell"`
	PriceSell      float64 `json:"priceSell"`
	ValueSell      float64 `json:"valueSell"`
	CommissionSell float64 `json:"commissionSell"`
	QuantityEnd    float64 `json:"quantityEnd"`
	PriceEnd       float64 `json:"priceEnd"`
	ValueEnd       float64 `json:"valueEnd"`
	DividendValue  float64 `json:"dividendValue"`
	DividendTax    float64 `json:"dividendTax"`
	CouponValue    float64 `json:"couponValue"`
	CouponTax      float64 `json:"couponTax"`
	TotalValue     float64 `json:"totalValue"`
	TotalPercent   float64 `json:"totalPercent"`
}

type Signal struct {
	Ticker     string
	Indicators []string
}

func (a *Analyser) Init(ivToken string) {

	a.Client = tinvestclient.Client{}

	a.Client.Init(ivToken)

}

func (a *Analyser) GetTotals(ivFrom time.Time, ivTo time.Time) (rsTotals Totals, roError error) {

	// ltOperations, loError := a.Client.GetOperations("", ivFrom, ivTo)

	// if loError != nil {
	// 	roError = loError
	// 	return
	// }

	// for _, lsOperation := range ltOperations {

	// 	switch lsOperation.Type {
	// 	case tinvestclient.OperationPayIn:

	// 	}

	// }

	return

}

func (a *Analyser) GetProfit(ivTicker string, ivFrom time.Time, ivTo time.Time) (rtProfit []Profit, roError error) {

	lvFigi := ""

	ltInstruments := make(map[string]tinvestclient.Instrument)

	if ivTicker == "" {

		ltInstruments, roError = a.GetInstruments()

		if roError != nil {
			return
		}

	} else {

		lsInstrument, loError := a.Client.GetInstrumentByTicker(ivTicker)

		if loError != nil {
			roError = loError
			return
		}

		ltInstruments[lsInstrument.FIGI] = lsInstrument

		lvFigi = lsInstrument.FIGI

	}

	ltOperationsAll, roError := a.GetOperations(lvFigi, ivFrom, ivTo)

	if roError != nil {
		return
	}

	for lvFIGI, ltOperations := range ltOperationsAll {

		lsInstrument := ltInstruments[lvFIGI]

		if lsInstrument.Type == tinvestclient.InstumentTypeCurrency {
			continue
		}

		ltCandles, loError := a.Client.GetCandles(lsInstrument.FIGI, tinvestclient.IntervalDay, ivTo.AddDate(0, 0, -15), ivTo)

		if loError != nil {
			roError = loError
			return
		}

		lsProfit := Profit{}

		lsProfit.Ticker = lsInstrument.Ticker
		lsProfit.Text = lsInstrument.Text
		lsProfit.Currency = lsInstrument.Currency

		for _, lsOperation := range ltOperations {

			switch lsOperation.Type {

			case tinvestclient.OperationBuy:

				lsProfit.QuantityBuy += lsOperation.Quantity
				lsProfit.ValueBuy += lsOperation.Value
				lsProfit.CommissionBuy += lsOperation.Commission

			case tinvestclient.OperationSell:

				lsProfit.QuantitySell += lsOperation.Quantity
				lsProfit.ValueSell += lsOperation.Value
				lsProfit.CommissionSell += lsOperation.Commission

			case tinvestclient.OperationDividend:
				lsProfit.DividendValue += lsOperation.Value

			case tinvestclient.OperationTaxDividend:
				lsProfit.DividendTax += lsOperation.Value

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
			lsProfit.DividendValue - lsProfit.DividendTax +
			lsProfit.CouponValue - lsProfit.CouponTax

		if lsProfit.ValueBuy != 0 {
			lsProfit.TotalPercent = lsProfit.TotalValue / lsProfit.ValueBuy * 100.
		}

		rtProfit = append(rtProfit, lsProfit)

	}

	return

}

func (a *Analyser) GetSignals(itTickers []string) (rtSignals []Signal, roError error) {

	ltInstruments, roError := a.GetInstruments()

	if roError != nil {
		return
	}

	for _, lvTicker := range itTickers {

		time.Sleep(100 * time.Millisecond)

		lsInstrument, lvExists := ltInstruments[lvTicker]

		if !lvExists {
			continue
		}

		ltCandles, loError := a.Client.GetCandles(lsInstrument.FIGI, tinvestclient.IntervalDay, time.Now().AddDate(0, 0, -60), time.Now())

		if loError != nil {
			roError = loError
			return
		}

		if len(ltCandles) == 0 {
			continue
		}

		lsSignal := Signal{
			Ticker: lsInstrument.Ticker,
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

func (a *Analyser) GetInstruments() (rtInstruments map[string]tinvestclient.Instrument, roError error) {

	ltInstruments, roError := a.Client.GetInstruments()

	if roError != nil {
		return
	}

	rtInstruments = make(map[string]tinvestclient.Instrument)

	for _, lsInstrument := range ltInstruments {
		rtInstruments[lsInstrument.FIGI] = lsInstrument
		rtInstruments[lsInstrument.Ticker] = lsInstrument
	}

	return

}

func (a *Analyser) GetOperations(ivFIGI string, ivFrom time.Time, ivTo time.Time) (rtOperations map[string][]tinvestclient.Operation, roError error) {

	ltOperations, roError := a.Client.GetOperations(ivFIGI, ivFrom, ivTo)

	if roError != nil {
		return
	}

	rtOperations = make(map[string][]tinvestclient.Operation)

	for _, lsOperation := range ltOperations {

		if lsOperation.FIGI == "" {
			continue
		}

		ltOperations := rtOperations[lsOperation.FIGI]

		ltOperations = append(ltOperations, lsOperation)

		rtOperations[lsOperation.FIGI] = ltOperations

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

		loCandle := techan.NewCandle(techan.TimePeriod{Start: lsCandle.Time, End: lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)
		loCandle.Volume = big.NewDecimal(lsCandle.Volume)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loMACDIndincator := techan.NewMACDIndicator(loClosePriceIndicator, 12, 26)

	loMACDHistogram := techan.NewMACDHistogramIndicator(loMACDIndincator, 9)

	lvPrevValue1 := loMACDHistogram.Calculate(len(itCandles) - 1).Float()
	lvPrevValue2 := loMACDHistogram.Calculate(len(itCandles) - 2).Float()
	lvPrevValue3 := loMACDHistogram.Calculate(len(itCandles) - 3).Float()

	if lvPrevValue1 < 0 &&
		lvPrevValue2 < 0 &&
		lvPrevValue3 < 0 &&
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

		loCandle := techan.NewCandle(techan.TimePeriod{Start: lsCandle.Time, End: lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)
		loCandle.Volume = big.NewDecimal(lsCandle.Volume)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loMACDIndincator := techan.NewMACDIndicator(loClosePriceIndicator, 12, 26)

	loMACDHistogram := techan.NewMACDHistogramIndicator(loMACDIndincator, 9)

	lvPrevValue1 := loMACDHistogram.Calculate(len(itCandles) - 1).Float()
	lvPrevValue2 := loMACDHistogram.Calculate(len(itCandles) - 2).Float()
	lvPrevValue3 := loMACDHistogram.Calculate(len(itCandles) - 3).Float()

	if lvPrevValue1 > 0 &&
		lvPrevValue2 > 0 &&
		lvPrevValue3 > 0 &&
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

		loCandle := techan.NewCandle(techan.TimePeriod{Start: lsCandle.Time, End: lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)
		loCandle.Volume = big.NewDecimal(lsCandle.Volume)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loRSIIndincator := techan.NewRelativeStrengthIndexIndicator(loClosePriceIndicator, 14)

	if loRSIIndincator.Calculate(len(itCandles)-1).Float() <= 30 {
		rvIs = true
	}

	return

}

func isRSISell(itCandles []tinvestclient.Candle) (rvIs bool) {

	rvIs = false

	loSeries := techan.NewTimeSeries()

	for _, lsCandle := range itCandles {

		loCandle := techan.NewCandle(techan.TimePeriod{Start: lsCandle.Time, End: lsCandle.Time})

		loCandle.OpenPrice = big.NewDecimal(lsCandle.Open)
		loCandle.ClosePrice = big.NewDecimal(lsCandle.Close)
		loCandle.MaxPrice = big.NewDecimal(lsCandle.High)
		loCandle.MinPrice = big.NewDecimal(lsCandle.Low)
		loCandle.Volume = big.NewDecimal(lsCandle.Volume)

		loSeries.AddCandle(loCandle)

	}

	loClosePriceIndicator := techan.NewClosePriceIndicator(loSeries)

	loRSIIndincator := techan.NewRelativeStrengthIndexIndicator(loClosePriceIndicator, 14)

	if loRSIIndincator.Calculate(len(itCandles)-1).Float() >= 70 {
		rvIs = true
	}

	return

}
