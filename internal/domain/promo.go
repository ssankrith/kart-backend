package domain

// PromoChecker validates coupon strings (e.g. file-based corpora).
type PromoChecker interface {
	Valid(code string) bool
	Close() error
}
