package main

type ReleaseDownload struct {
	ReleaseTag string `json:"release_tag"`
	Download   string `json:"download"`
}

type Downloads struct {
	PageProps PageProps `json:"pageProps"`
	NSsg      bool      `json:"__N_SSG"`
}
type Assets struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OS          string   `json:"os"`
	Arch        string   `json:"arch"`
	Size        int      `json:"size"`
	Sha256      string   `json:"sha256"`
	PublicURL   string   `json:"publicUrl"`
	ReleaseIds  []string `json:"releaseIds"`
}
type Versions struct {
	ReleaseID string   `json:"releaseId"`
	Product   string   `json:"product"`
	Version   string   `json:"version"`
	NotesMd   string   `json:"notesMd"`
	Status    string   `json:"status"`
	Assets    []Assets `json:"assets"`
}
type InitialDownloads struct {
	MajorVersion string     `json:"majorVersion"`
	Versions     []Versions `json:"versions"`
}
type Teleport struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Os          string   `json:"os"`
	Arch        string   `json:"arch"`
	Size        int      `json:"size"`
	Sha256      string   `json:"sha256"`
	PublicURL   string   `json:"publicUrl"`
	ReleaseIds  []string `json:"releaseIds"`
}
type RecommendedDownloads struct {
	Version   string      `json:"version"`
	Teleport  Teleport    `json:"teleport"`
	TshClient interface{} `json:"tshClient"`
	Connect   interface{} `json:"connect"`
}
type PageProps struct {
	Os                   string               `json:"os"`
	InitialDownloads     []InitialDownloads   `json:"initialDownloads"`
	RecommendedDownloads RecommendedDownloads `json:"recommendedDownloads"`
}
