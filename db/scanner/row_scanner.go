package scanner

type RowScanner interface {
	Scan(dest ...any) error
}
