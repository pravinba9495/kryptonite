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
	lastBuyPrice     float64
	isTriggered      bool
}

func (pm *PriceMonitor) SwitchOrderType(orderType OrderType, price float64) {
	pm.isTriggered = false

	if orderType == SellOrder {
		pm.currentOrderType = SellOrder
		pm.lastBuyPrice = price
		pm.triggerPriceUp = price * (1 + (pm.limitPercent / 100))
		pm.triggerPriceDown = price * (1 - (pm.stopLossPercent / 100))
	}

	if orderType == BuyOrder {
		pm.currentOrderType = BuyOrder
		pm.lastBuyPrice = 0
		pm.triggerPriceUp = price * (1 + (pm.stopLossPercent / 100))
		pm.triggerPriceDown = price * (1 - (pm.limitPercent / 100))
	}
}

func (pm *PriceMonitor) IsTriggered() bool {
	return pm.isTriggered
}

func (pm *PriceMonitor) Update(currentPrice float64) {
	isTriggered := false

	if pm.currentOrderType == BuyOrder {
		pm.lastBuyPrice = 0
		if currentPrice >= pm.triggerPriceUp {
			isTriggered = true
			pm.SwitchOrderType(SellOrder, currentPrice)
		} else {
			newTriggerPriceDown := currentPrice * (1 - (pm.limitPercent / 100))
			if newTriggerPriceDown < pm.triggerPriceDown {
				pm.triggerPriceDown = newTriggerPriceDown
			}

			newTriggerPriceUp := currentPrice * (1 + (pm.stopLossPercent / 100))
			if newTriggerPriceUp < pm.triggerPriceUp {
				pm.triggerPriceUp = newTriggerPriceUp
			}
		}
	} else if pm.currentOrderType == SellOrder {
		if currentPrice <= pm.triggerPriceDown {
			isTriggered = true
			pm.SwitchOrderType(BuyOrder, currentPrice)
		} else {
			newTiggerPriceUp := currentPrice * (1 + (pm.limitPercent / 100))
			if newTiggerPriceUp > pm.triggerPriceUp {
				pm.triggerPriceUp = newTiggerPriceUp
			}

			newTriggerPriceDown := currentPrice * (1 - (pm.stopLossPercent / 100))
			if newTriggerPriceDown > pm.triggerPriceDown {
				pm.triggerPriceDown = newTriggerPriceDown
			}
		}

	} else {
		panic("unknown order type")
	}

	pm.isTriggered = isTriggered
}

func NewPriceMonitor(initialOrderType OrderType, initialPrice float64, lastBuyPrice float64, limitPercent float64, stopLossPercent float64) *PriceMonitor {
	var pm PriceMonitor

	if initialOrderType == BuyOrder {
		pm = PriceMonitor{
			currentOrderType: initialOrderType,
			limitPercent:     limitPercent,
			triggerPriceUp:   initialPrice * (1 + (limitPercent / 100)),
			triggerPriceDown: initialPrice * (1 - (limitPercent / 100)),
			stopLossPercent:  stopLossPercent,
			lastBuyPrice:     0,
			isTriggered:      false,
		}
	} else if initialOrderType == SellOrder {
		pm = PriceMonitor{
			currentOrderType: initialOrderType,
			limitPercent:     limitPercent,
			triggerPriceUp:   lastBuyPrice * (1 + (limitPercent / 100)),
			triggerPriceDown: lastBuyPrice * (1 - (stopLossPercent / 100)),
			stopLossPercent:  stopLossPercent,
			lastBuyPrice:     lastBuyPrice,
			isTriggered:      false,
		}
	} else {
		panic("unknown order type")
	}

	return &pm
}
