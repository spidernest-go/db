package mongo

import (
	"fmt"
	"math/rand"
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"upper.io/db"
)

const (
	testRows = 1000
)

func updatedArtistN(i int) string {
	return fmt.Sprintf("Updated Artist %d", i%testRows)
}

func artistN(i int) string {
	return fmt.Sprintf("Artist %d", i%testRows)
}

func connectAndAddFakeRows() (db.Database, error) {
	var err error
	var sess db.Database

	if sess, err = db.Open(Adapter, settings); err != nil {
		return nil, err
	}

	driver := sess.Driver().(*mgo.Session)

	if err = driver.DB("").C("artist").DropCollection(); err != nil {
		return nil, err
	}

	for i := 0; i < testRows; i++ {
		if err = driver.DB("").C("artist").Insert(bson.M{"name": artistN(i)}); err != nil {
			return nil, err
		}
	}

	return sess, nil
}

func BenchmarkMgoAppend(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = db.Open(Adapter, settings); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	driver := sess.Driver().(*mgo.Session)

	if err = driver.DB("").C("artist").DropCollection(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = driver.DB("").C("artist").Insert(bson.M{"name": "Hayao Miyazaki"}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpperAppend benchmarks an insertion by upper.io/db.
func BenchmarkUpperAppend(b *testing.B) {

	sess, err := db.Open(Adapter, settings)
	if err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	artist, err := sess.Collection("artist")
	if err != nil {
		b.Fatal(err)
	}

	artist.Truncate()

	item := struct {
		Name string `bson:"name"`
	}{"Hayao Miyazaki"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = artist.Append(item); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpperAppendVariableArgs benchmarks an insertion by upper.io/db
// with variable parameters.
func BenchmarkUpperAppendVariableArgs(b *testing.B) {

	sess, err := db.Open(Adapter, settings)
	if err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	artist, err := sess.Collection("artist")
	if err != nil {
		b.Fatal(err)
	}

	artist.Truncate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := struct {
			Name string `bson:"name"`
		}{fmt.Sprintf("Hayao Miyazaki %d", rand.Int())}
		if _, err = artist.Append(item); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMgoSelect benchmarks MongoDB find queries.
func BenchmarkMgoSelect(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	driver := sess.Driver().(*mgo.Session)

	type artistType struct {
		Name string `bson:"name"`
	}

	var item artistType

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = driver.DB("").C("artist").Find(bson.M{"name": artistN(i)}).One(&item); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMgoSelect benchmarks MongoDB find queries.
func BenchmarkMgoSelectAll(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	driver := sess.Driver().(*mgo.Session)

	type artistType struct {
		Name string `bson:"name"`
	}

	var items []artistType

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = driver.DB("").C("artist").Find(bson.M{"name": artistN(i)}).All(&items); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpperFind benchmarks upper.io/db's One method.
func BenchmarkUpperFind(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	artist, err := sess.Collection("artist")
	if err != nil {
		b.Fatal(err)
	}

	type artistType struct {
		Name string `bson:"name"`
	}

	var item artistType

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := artist.Find(db.Cond{"name": artistN(i)})
		if err = res.One(&item); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpperFindAll benchmarks upper.io/db's All method.
func BenchmarkUpperFindAll(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	artist, err := sess.Collection("artist")
	if err != nil {
		b.Fatal(err)
	}

	type artistType struct {
		Name string `bson:"name"`
	}

	var items []artistType

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := artist.Find(db.Or{
			db.Cond{"name": artistN(i)},
			db.Cond{"name": artistN(i + 1)},
			db.Cond{"name": artistN(i + 2)},
		})
		if err = res.All(&items); err != nil {
			b.Fatal(err)
		}
		if len(items) != 3 {
			b.Fatal("Expecting 3 results.")
		}
	}
}

// BenchmarkMgoUpdate benchmarks MongoDB update queries.
func BenchmarkMgoUpdate(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	driver := sess.Driver().(*mgo.Session)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = driver.DB("").C("artist").UpdateAll(bson.M{"name": artistN(i)}, bson.M{"$set": bson.M{"name": updatedArtistN(i)}}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpperUpdate benchmarks upper.io/db's Update method.
func BenchmarkUpperUpdate(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	artist, err := sess.Collection("artist")
	if err != nil {
		b.Fatal(err)
	}

	type artistType struct {
		Name string `bson:"name"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newValue := artistType{
			Name: updatedArtistN(i),
		}
		res := artist.Find(db.Cond{"name": artistN(i)})
		if err = res.Update(newValue); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMgoDelete benchmarks MongoDB delete queries.
func BenchmarkMgoDelete(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	driver := sess.Driver().(*mgo.Session)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = driver.DB("").C("artist").RemoveAll(bson.M{"name": artistN(i)}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpperRemove benchmarks
func BenchmarkUpperRemove(b *testing.B) {
	var err error
	var sess db.Database

	if sess, err = connectAndAddFakeRows(); err != nil {
		b.Fatal(err)
	}

	defer sess.Close()

	artist, err := sess.Collection("artist")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := artist.Find(db.Cond{"name": artistN(i)})
		if err = res.Remove(); err != nil {
			b.Fatal(err)
		}
	}
}