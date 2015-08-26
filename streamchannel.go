package eventual2go

type streamchannel chan Data

func (sc streamchannel) pipe(t chan Data) {
	pile := []interface{}{}
	ok := true
	for ok {
		if len(pile) == 0 {
			pile = append(pile, <-sc)
		} else {
			select {
			case t <- pile[0]:
				pile = pile[1:]
			case d, ok := <-sc:
				if !ok {
					return
				}
				pile = append(pile, d)

			}
		}
	}
}

