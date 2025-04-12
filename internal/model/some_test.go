package model

import "testing"

type RecordMarshaller[T any] interface {
	MarshallRecord(T) []string
	UnmarshallRecord([]string) (T, error) 
}

type user struct {
	ID int
	Name string
}

type UserMarshaller struct {}

func (_ UserMarshaller) MarshallRecord(u user) []string {
	return []string{"marshalled", "user"}
}

func (_ UserMarshaller) UnmarshallRecord(s []string) (user, error){
	return user{}, nil 
}

type session struct {
	ID int
	Age int 
}

type SessionMarshaller struct {}

func (_ SessionMarshaller) MarshallRecord(s session) []string {
	return []string{"marshalled", "user"}
}

func (_ SessionMarshaller) UnmarshallRecord(s []string) (session, error){
	return session{}, nil 
}

func TestUserMarshaller(t *testing.T) {
	var um UserMarshaller
	record := marshal(um, user{})

	if len(record) != 2 {
		t.Fatal("Len must be two")
	}

	var sm SessionMarshaller
	record = marshal(sm, session{})

	if len(record) != 2 {
		t.Fatal("Len must be two")
	}
}

func marshal[T any](rm RecordMarshaller[T], t T) []string{
	return rm.MarshallRecord(t)	
}
