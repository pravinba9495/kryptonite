package main

type OrderType int

const (
	BuyOrder OrderType = iota
	SellOrder
)

var orderTypes = map[OrderType]string{
	BuyOrder:  "BUY",
	SellOrder: "SELL",
}

func (ot OrderType) String() string {
	return orderTypes[ot]
}

type PriceMonitor struct {
	currentOrderType OrderType
	limitPercent     float64
	triggerPriceUp   float64
	triggerPriceDown float64
	stopLossPercent  float64
	previousPrice    float64
	isTriggered      bool
}

func (pm *PriceMonitor) SwitchOrderType(orderType OrderType, triggerPriceUp float64, triggerPriceDown float64) {
	pm.currentOrderType = orderType
	pm.previousPrice = 0
	pm.triggerPriceUp = 1e10
	pm.triggerPriceDown = -1
	pm.isTriggered = false

	if triggerPriceUp > 0 {
		pm.triggerPriceUp = triggerPriceUp
	}

	if triggerPriceDown > 0 {
		pm.triggerPriceDown = triggerPriceDown
	}
}

func (pm *PriceMonitor) IsTriggered() bool {
	return pm.isTriggered
}

func (pm *PriceMonitor) Update(currentPrice float64) {
	isTriggered := false

	if pm.currentOrderType == BuyOrder {
		if currentPrice >= pm.triggerPriceUp || currentPrice <= pm.triggerPriceDown {
			isTriggered = true
		} else {
			newTriggerPriceDown := currentPrice * (1 - (pm.limitPercent / 100))
			if newTriggerPriceDown < pm.triggerPriceDown || pm.triggerPriceDown < 0 {
				pm.triggerPriceDown = newTriggerPriceDown
			}

			newTriggerPriceUp := currentPrice * (1 + (pm.stopLossPercent / 100))
			if newTriggerPriceUp < pm.triggerPriceUp {
				pm.triggerPriceUp = newTriggerPriceUp
			}
		}
	} else if pm.currentOrderType == SellOrder {
		if currentPrice >= pm.triggerPriceUp || currentPrice <= pm.triggerPriceDown {
			isTriggered = true
		} else {
			newTiggerPriceUp := currentPrice * (1 + (pm.limitPercent / 100))
			if newTiggerPriceUp > pm.triggerPriceUp || pm.triggerPriceUp > 1e9 {
				pm.triggerPriceUp = newTiggerPriceUp
			}

			newTriggerPriceDown := currentPrice * (1 - (pm.stopLossPercent / 100))
			if newTriggerPriceDown > pm.triggerPriceDown || pm.triggerPriceDown < 0 {
				pm.triggerPriceDown = newTriggerPriceDown
			}
		}

	} else {
		panic("unknown order type")
	}

	pm.previousPrice = currentPrice
	pm.isTriggered = isTriggered
}

func NewPriceMonitor(initialOrderType OrderType, initialPrice float64, lastBuyPrice float64, limitPercent float64, stopLossPercent float64) *PriceMonitor {
	var pm PriceMonitor

	if initialOrderType == BuyOrder {
		pm = PriceMonitor{
			currentOrderType: initialOrderType,
			limitPercent:     limitPercent,
			triggerPriceUp:   initialPrice * (1 + (stopLossPercent / 100)),
			triggerPriceDown: initialPrice * (1 - (limitPercent / 100)),
			stopLossPercent:  stopLossPercent,
			isTriggered:      false,
		}
	} else if initialOrderType == SellOrder {
		pm = PriceMonitor{
			currentOrderType: initialOrderType,
			limitPercent:     limitPercent,
			triggerPriceUp:   lastBuyPrice * (1 + (limitPercent / 100)),
			triggerPriceDown: lastBuyPrice * (1 - (stopLossPercent / 100)),
			stopLossPercent:  stopLossPercent,
			isTriggered:      false,
		}
	} else {
		panic("unknown order type")
	}

	return &pm
}
