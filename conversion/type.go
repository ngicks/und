package conversion

import "fmt"

type Empty []struct{}

func (e Empty) MarshalJSON() ([]byte, error) {
	return []byte(`null`), nil
}

func (e *Empty) UnmarshalJSON(data []byte) error {
	if string(data) != "null" {
		return fmt.Errorf("Empty: must only be null but is %s", data)
	}
	return nil
}
