package csv

import (
	"time"
)

type KEY struct {
	measurement string
	fkeys       []string
	field       string
	time        time.Time
}

type VALUE struct {
	fieldVals []string
	value     string
}

// Cmp compares a and b. Return value is:
//
//	< 0 if a <  b
//	  0 if a == b
//	> 0 if a >  b
//
func CmpKEY(a, b KEY) int {
	if a.measurement < b.measurement {
		return -1
	}
	if a.measurement > b.measurement {
		return 1
	}
	return 0
}

// type fluxTableUniqueness struct {
// 	measurement string
// 	fkeys       []string
// 	field       string
// }

// type tableUniquenesses []fluxTableUniqueness

// func (tu tableUniquenesses) Len() int {
// 	return len(tu)
// }

// func (tu tableUniquenesses) Less(i, j int) bool {
// 	if tu[i].measurement < tu[j].measurement {
// 		return true
// 	}
// 	if tu[i].measurement > tu[j].measurement {
// 		return false
// 	}
// 	return tu[i].field < tu[j].field
// }
// func (tu tableUniquenesses) Swap(i, j int) {
// 	temp := tu[i]
// 	tu[i] = tu[j]
// 	tu[j] = temp
// }

// type fluxRowUniqueness struct {
// 	time  time.Time
// 	value string
// }

// type rowUniquenesses []fluxRowUniqueness

// func (tu rowUniquenesses) Len() int {
// 	return len(tu)
// }

// func (tu rowUniquenesses) Less(i, j int) bool {
// 	return tu[i].time.Before(tu[j].time)
// }

// func (tu rowUniquenesses) Swap(i, j int) {
// 	temp := tu[i]
// 	tu[i] = tu[j]
// 	tu[j] = temp
// }

// func toCSV(writer io.Writer, start, stop time.Time, rm ...lp.Metric) error {
// 	tables := tableUniquenesses{}
// 	for i := range rm {
// 		for j := range rm[i].FieldList {
// 			table := fluxTableUniqueness{
// 				measurement: rm[i].Name(),
// 				field:       *(rm[i].FieldList()[j]),
// 				tags:        *(rm[i].TagList()[j]),
// 			}
// 		}
// 	}
// 	return nil
// }
