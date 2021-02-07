package newsfeed

type Getter interface {
	GetAll() []Item
}

type Adder interface {
	Add(item Item)
}

type Item struct {
	Title string `json:"title"`
	Post  string `json:"post"`
}

// Repo is type of object consisting of slice with name 'Items' and type []Item
type Repo struct {
	Items []Item
}

//struct Repo pointer as a return type
func New() *Repo {
	return &Repo{ // return address of Repo struct
		Items: []Item{},
	}
}

func (r *Repo) Add(item Item) {
	r.Items = append(r.Items, item)
}

func (r *Repo) GetAll() []Item {
	return r.Items
}
