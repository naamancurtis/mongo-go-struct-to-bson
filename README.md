<h1> Mongo Go - Struct to BSON </h1>

Provides utility methods to support the converting of `structs` to `bson maps` for use in various MongoDB queries/patch updates. It is intended to be used alongside the [Mongo-Go Driver](https://github.com/mongodb/mongo-go-driver)

Essentially it will convert a `struct` to `bson.M`

<h3> Original Source </h3>

This was built by stripping down and modifying **Fatih Arslan** & **Cihangir Savas's** [Structs](https://github.com/fatih/structs) package.

<h2> Contents </h2>

- [Installation](#installation)
- [Basic Usage & Examples](#basic-usage--examples)
  - [Calling ConvertStructToBSONMap with Options](#calling-convertstructtobsonmap-with-options)
    - [Examples](#examples)
  - [Using a different Tag Name](#using-a-different-tag-name)
- [Known Issues](#known-issues)
  - [Zero Values](#zero-values)
- [Getting involved](#getting-involved)

### Installation

```sh
go get https://github.com/naamancurtis/mongo-go-struct-to-bson
```

### Basic Usage & Examples

To get started quickly simply import the package and set up the struct you want to be converted.

```go
type User struct {
  ID        primitive.ObjectID `bson:"_id"`
  FirstName string `bson:"firstName"`

  // "omitempty" fields will not exist in the bson.M if they hold a zero value
  LastName  string `bson:"lastName,omitempty"`

  // "string" fields will hold the string value as defined by their implementation of the Stringer() interface
  DoB       time.Time `bson:"dob,string"`

  // "flatten" fields will have all of the fields within the nested struct moved up one level in the map to sit at the higher level
  Characteristics *Characteristics `bson:"characteristics,flatten"`

  // "omitnested" fields will not be recursively mapped, instead it will just hold the value
  Metadata  Metadata `bson:"metadata,omitnested"`

  // "-" fields will not be mapped at all
  Secret string `bson:"-"`

  // Unexported fields will not be mapped
  favouriteColor string
}

type Characteristics struct {
  LeftHanded bool `bson:"leftHanded"`
  Tall bool `bson:"tall"`
}

type Metadata struct {
  LastActive time.Time `bson:"lastActive"`
}
```

Then in order to map the struct you can use the convience method `ConvertStructToBSONMap()`, this handles the interim step of converting your struct to a `StructToBSON` struct _(which is used as a wrapper to provide the relevant methods)_ and then converts the wrapped struct, returning a `bson.M`

```go
user = User {
  ID:              "54759eb3c090d83494e2d804", // would actually hold the primitive.ObjectID value
  FirstName:       "Jane",
  LastName:        "",
  DoB:             time.Date(1985,6,15,0,0,0,0,time.UTC),
  Characteristics: &Characteristics{
    LeftHanded: true,
    Tall:       false,
  },
  Metadata:        Metadata{LastActive: time.Date(2019, 7,23, 14,0,0,0,time.UTC)},
  Secret:          "secret",
  favouriteColor:  "blue",
}

// Calling the Convert function - passing nil as the options for now.
result := ConvertStructToBSONMap(user, nil)

// The result would be:
bson.M {
  "_id": "54759eb3c090d83494e2d804", // would actually hold the primitive.ObjectID value
  "firstName": "Jane",
  "dob": "1985-06-15 00:00:00 +0000 UTC",
  "leftHanded": true,
  "tall": false,
  "metadata": Metadata {
    LastActive: time.Date(2019, 7,23, 14,0,0,0,time.UTC)
  },

  // Notes:
  // - dob: holds the Stringer() representation of time.Time
  // - Characterstics: struct has been flattened, so it's values are one level up
  // - Metadata: had the omitnested tag, so holds the actual value of the struct, not a bson.M
  // - lastName: has been omitted as it held the zero value for a string
  // - secret & favouriteColor: have been omitted, as they either had the "-" tag or were unexported
}
```

#### Calling ConvertStructToBSONMap with Options

There are currently 3 options available to pass to `ConvertStructToBSONMap()`, they're all held in a `MappingOpts` struct and default to a value of `false` if they're either unset or a value of `nil` is used as `MappingOpts`.

1. `UseIDifAvailable` - Will just return `bson.M { "_id": idVal }` if the _"\_id"_ tag is present in that struct, if it is not present or holds a zero value it will map the struct as you would expect. This flag has priority over the other 3 options.
2. `RemoveID` - Will remove any _"\_id"_ fields from your `bson.M`
3. `GenerateFilterOrPatch` - If true, it will check all struct fields for zero type values and omit any that are found regardless of any tag options, effectively it enforces the behaviour of the `"omitempty"` tag, regardless of whether the struct field has it or not

##### Examples

```go
// Using the same struct as the examples above

user = User {
  ID:              "54759eb3c090d83494e2d804",
  FirstName:       "Jane",
  LastName:        "",
  DoB:             time.Date(1985,6,15,0,0,0,0,time.UTC),
  Characteristics: &Characteristics{
    LeftHanded: true,
    Tall:       false,
  },
  Metadata:        Metadata{LastActive: time.Date(2019, 7, 23, 14, 0, 0, 0, time.UTC)},
  Secret:          "secret",
  favouriteColor:  "blue",
}
```

<h6>UseIDifAvailable = true</h6>

```go
result := ConvertStructToBSONMap(user, &MappingOpts{UseIDifAvailable: true})

// result would be:
bson.M { "\_id": "54759eb3c090d83494e2d804" }
// Value is indicative - it would actually hold the primitive.ObjectID value
```

<h6>RemoveID = true</h6>

```go
result := ConvertStructToBSONMap(user, &MappingOpts{RemoveID: true})

// result would be:
bson.M {
  "firstName": "Jane",
  "dob": "1985-06-15 00:00:00 +0000 UTC",
  "leftHanded": true,
  "tall": false,
  "metadata": Metadata{LastActive: time.Date(2019, 7, 23, 14, 0, 0, 0, time.UTC)},
}
```

<h6>GenerateFilterOrPatch = true</h6>

```go
// Using a modified user:
user = User {
  FirstName:       "Jane",
  LastName:        "",
  Characteristics: &Characteristics{
    LeftHanded: true,
    Tall:       false,
  },
  favouriteColor:  "blue",
}

result := ConvertStructToBSONMap(user, &MappingOpts{GenerateFilterOrPatch: true})

// result would be:
bson.M {
  "firstName": "Jane",
  "leftHanded": true,
}

// Note on the behaviour here: As go uses zero values, lastName & Characteristics.Tall
// are ignored, as they hold zero values.
// Please be aware of this when using this package
```

#### Using a different Tag Name

By default, the mapper uses the `"bson"` tag to identify what options and names should be assigned to each struct field. It is possible to change this to a custom tag if desired. In order to do so you need to split up the creation and mapping of your struct:

```go
// 1. Create the struct wrapper
tempStruct := NewBSONMapperStruct(myStruct)

// 2. Set the custom tag name
tempStruct.SetTagName("customTag")

// 3. Convert the struct to bson.M
result := tempStruct.ToBSONMap(nil) // Passing nil as the options in this example
```

### Known Issues

#### Zero Values

When using either the `omitempty` tag or `MappingOpts{GenerateFilterOrPatch: true}` and zero values mean something in your struct, ie. `""` or `false` **are** valid, meaningful values for a field. You will have to work around the fact that they **won't** be included in your returned `bson.M` as they are identified as _zero values_ and removed from the map.

One option is to add them manually to the returned bson.M once the mapping has occured using any checks you need to perform specific to your use case.

At the moment I can't work out an approach around this, however if anyone has any ideas then I'm all ears.

### Getting involved

If you'd like to get involved or contribute to this package, please feel free to do so, whether it's recomendations, code improvements or additional functionality.

When making any changes, please make sure to fix/add any tests prior to submitting the PR - ideally I'd like to keep the test coverage to as close as 100% as possible.
