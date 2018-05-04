package web

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/HALtheWise/o-links/context"
)

// Regression test
func TestBlankGenerator(t *testing.T) {
	e := needEnv(t)
	defer e.destroy(t)

	oldnouns := nouns
	oldadjectives := adjectives

	defer func() {
		nouns = oldnouns
		adjectives = oldadjectives
	}()

	nouns = strings.Split("cat dog", " ")
	adjectives = strings.Split("small large", " ")

	randsource = rand.New(rand.NewSource(42))

	desired := []string{"large dog", "small cat", "large cat", "small dog", "small small dog", "large small cat", "large small dog", "small small cat", "large large cat", "small large dog", "large large dog", "generated-5", "small large cat", "generated-7", "generated-2", "generated-6"}

	var results []string

	for i := range desired {
		uid := fmt.Sprint(rand.Uint64())
		link, err := generateLink(e.ctx, uid)
		if err != nil {
			t.Fatalf("Unable to generate link #%d, %s", i+1, err)
		}
		results = append(results, link)

		err = e.ctx.Put(link, &context.Route{Uid: uid, URL: "https://google.com"})
		if err != nil {
			t.Fatalf("Unable to put route in database: %s", err)
		}
	}

	if !reflect.DeepEqual(desired, results) {
		t.Errorf("Wrong sequence of links generated: Expected \n%#v\n got \n%#v",
			desired, results)
	}
}
