package payout

// PayoutItem represents a commission included in a payout.
// This is a value object - immutable once created.
type PayoutItem struct {
	commissionID uint
	orderID      string
	amount       float64
}

// NewPayoutItem creates a new PayoutItem.
func NewPayoutItem(commissionID uint, orderID string, amount float64) PayoutItem {
	return PayoutItem{
		commissionID: commissionID,
		orderID:      orderID,
		amount:       amount,
	}
}

// Getters
func (i PayoutItem) CommissionID() uint { return i.commissionID }
func (i PayoutItem) OrderID() string    { return i.orderID }
func (i PayoutItem) Amount() float64    { return i.amount }

// Equals compares two payout items.
func (i PayoutItem) Equals(other PayoutItem) bool {
	return i.commissionID == other.commissionID
}
