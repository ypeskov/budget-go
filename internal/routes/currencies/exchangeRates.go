package currencies

func calcAmountInBaseCurrency(amount float64, exchangeRate float64) float64 {
	return amount * exchangeRate
}
