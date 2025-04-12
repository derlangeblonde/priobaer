package model

type DbError struct {
	MsgForUser   string
	Err error
}

func DefaultDbError(err error) DbError {
	return DbError{
		MsgForUser:   "Datenbankfehler",
		Err: err,
	}
}

func (d DbError) Error() string {
	return d.MsgForUser + ": " + d.Err.Error()
}
