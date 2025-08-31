package domain

// LoanRepository defines the interface for loan data persistence
type LoanRepository interface {
	Create(loan *Loan) error
	GetByID(id string) (*Loan, error)
	Update(loan *Loan) error
	List() ([]*Loan, error)
	Delete(id string) error
}
