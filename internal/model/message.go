package model

type Message struct {
	Data     Data
	UserFrom string
	ChatFrom Chat
	ChatTo   Chat
}

type Data struct {
	Text   string
	Photos []string
	Gif    []string
	Doc    []string
}

func (d *Data) AddGif(url string) {
	if url != "" {
		d.Gif = append(d.Gif, url)
	}
}

func (d *Data) AddFile(url string) {
	if url != "" {
		d.Doc = append(d.Doc, url)
	}
}

func (d *Data) AddPhoto(url string) {
	if url != "" {
		d.Photos = append(d.Photos, url)
	}
}
