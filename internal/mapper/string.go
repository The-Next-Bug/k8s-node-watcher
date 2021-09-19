package mapper

import (
	"fmt"
)

func (sm *serverMapping) String() string {
	return fmt.Sprintf("%+v", *sm)
}
