package credit

import "context"

type CreditSrv interface {
	Debit(ctx context.Context, orgId string, amount int) (newBalance int, err error)
}
