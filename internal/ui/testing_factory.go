package ui

import "softbaer.dev/ass/internal/seededuuid"

func RandomParticipant() Participant {
	prename := seededuuid.SeededUUID()
	surname := seededuuid.SeededUUID()

	p := Participant{Prename: prename.String(), Surname: surname.String()}

	return p
}

func RandomParticipants(n int) (result []Participant) {
	for i := 0; i < n; i++ {
		result = append(result, RandomParticipant())
	}

	return
}
