package provider

type EmailSender interface {
	Send(
		email string,
		orderID string,
		amount int64,
	) error
}