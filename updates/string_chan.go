package updates

/** This is a template of an easy to use channel with NON-blocking Push and blocking Get
 *  It can be activated or not, an unactivated channel is doing nothing
 *	and Get(s) are still blocking (and never resolves)
 *
 *  Only T, defaultValue, interface name and function name NewXxxChan needs to be change for easy reuse
 */

/*T .*/
type T = string

const defaultValue = ""

/*StringChan .*/
type StringChan interface {
	Get() T
	Push(file T)
}

type matchChan struct {
	Chan
}

/*NewStringChan .*/
func NewStringChan(activated bool) StringChan {
	return &matchChan{Chan: NewChan(activated)}
}

func (ch *matchChan) Push(match T) {
	ch.Chan.Push(match)
}

func (ch *matchChan) Get() T {
	match, ok := ch.Chan.Get().(T)
	if !ok {
		return defaultValue
	}
	return match
}
