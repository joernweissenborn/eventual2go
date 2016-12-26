package typed_events


//go:generate event_generator -e -t string
//go:generate event_generator -e -t int
//go:generate event_generator -e -t bool
//go:generate event_generator -e -t error
//go:generate event_generator -t []string -n StringSlice
