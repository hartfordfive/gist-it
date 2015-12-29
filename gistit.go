package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_SETTINGS_FILE = "~/.gistit"
)

var api_methods map[string]string

type GistIt struct {
	GitClient *github.Client
	User      string
	Debug     bool
}

func main() {

	app := cli.NewApp()
	app.Name = "Gist-it"

	gi := NewGistIt()

	fmt.Println()

	app.Commands = []cli.Command{

		{
			Name:  "create",
			Usage: "gist-it create ",
			Action: func(c *cli.Context) {

				if len(c.Args()) == 0 {
					errorAndExit("You must specify at least one file")
				}

				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter description:\n")
				description, _ := reader.ReadString('\n')
				fmt.Print("Is public (y/n):\n")
				is_public, _ := reader.ReadString('\n')
				gi.Create(description, is_public, c.Args())
			},
		},
		{
			Name:  "list",
			Usage: "gist-it list ",
			Action: func(c *cli.Context) {
				gi.MyList(c.Args())
			},
		},
		{
			Name:  "get",
			Usage: "gist-it get ",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 0 {
					errorAndExit("You must specify the gist ID")
				}
				gi.Get(c.Args())
			},
		},
	}
	app.Run(os.Args)
}

func NewGistIt() *GistIt {

	conf, err := loadSettings(DEFAULT_SETTINGS_FILE)
	if err != nil {
		fmt.Println("Error: Could not load settings file")
		fmt.Println(err)
		os.Exit(1)
	}
	debug, _ := strconv.ParseBool(conf["debug"])

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		token = conf["token"]
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return &GistIt{
		GitClient: github.NewClient(tc),
		User:      conf["user"],
		Debug:     debug,
	}

}

func (g *GistIt) MyList(params []string) {

	gists, _, err := g.GitClient.Gists.List(g.User, &github.GistListOptions{
		Since: time.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
	})

	if err != nil {
		errorAndExit(err.Error())
	}

	if len(gists) == 0 {
		msgAndExit("No gists to display")
	}

	fmt.Println(strings.Repeat("-", 148))
	fmt.Printf("|%s|%s|%s|%s|%s|%s|\n",
		centerString("ID", 26),
		centerString("Public", 10),
		centerString("Created", 13),
		centerString("Updated", 13),
		centerString("URL", 50),
		centerString("Desc", 28))
	fmt.Println(strings.Repeat("-", 148))

	for _, g := range gists {
		desc := *g.Description
		fmt.Printf("|%s|%v|%v|%v|%s|%s|\n",
			centerString(*g.ID, 26),
			centerString(strconv.FormatBool(*g.Public), 10),
			centerString(g.CreatedAt.Format("2006-01-02"), 13),
			centerString(g.UpdatedAt.Format("2006-01-02"), 13),
			centerString(*g.HTMLURL, 50),
			centerString(desc, 28))
		fmt.Println(strings.Repeat("-", 148))
	}

	fmt.Println()
	os.Exit(0)
}

func (g *GistIt) Create(description string, is_public string, files []string) {

	// Last arg is the description
	//

	var public bool
	if strings.ToLower(is_public) == "y" {
		public = true
	} else {
		public = false
	}

	var gist_files map[github.GistFilename]github.GistFile = map[github.GistFilename]github.GistFile{}

	for _, f := range files {
		data := getFileContents(f)
		gist_files[github.GistFilename(f)] = github.GistFile{
			Content: &data,
		}
	}

	gist, _, err := g.GitClient.Gists.Create(&github.Gist{
		Description: &description,
		Public:      &public,
		Files:       gist_files,
	})

	if err != nil {
		errorAndExit(err.Error())
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(centerString("Gist Details", 80))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(fmt.Sprintf("%15s", "ID:"), *gist.ID)
	fmt.Println(fmt.Sprintf("%15s", "Description:"), *gist.Description)
	fmt.Println(fmt.Sprintf("%15s", "Public:"), strconv.FormatBool(*gist.Public))
	fmt.Println(fmt.Sprintf("%15s", "URL:"), *gist.HTMLURL)
	fmt.Println(fmt.Sprintf("%15s", "Files:"), strings.Join(files, ","))
	fmt.Println(fmt.Sprintf("%15s", "Total Files:"), strconv.Itoa(len(gist.Files)))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()
	os.Exit(0)

}

func (g *GistIt) Get(params []string) {
	gist, _, err := g.GitClient.Gists.Get(params[0])
	if err != nil {
		errorAndExit(err.Error())
	}

	fmt.Println(centerString("Gist Details", 80))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(fmt.Sprintf("%15s", "ID:"), *gist.ID)
	fmt.Println(fmt.Sprintf("%15s", "Description:"), *gist.Description)
	fmt.Println(fmt.Sprintf("%15s", "Public:"), strconv.FormatBool(*gist.Public))
	// Save content to each files locally
	files := []string{}
	for fname, details := range gist.Files {
		files = append(files, string(fname))
		writeToFile(fmt.Sprintf("%s_%s", *gist.ID, fname), []byte(*details.Content))
	}
	fmt.Println(fmt.Sprintf("%15s", "Files:"), strings.Join(files, ","))
	fmt.Println(fmt.Sprintf("%15s", "Total Files:"), strconv.Itoa(len(gist.Files)))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Saved files have been prefixed with \"%s_\"\n", *gist.ID)
	fmt.Println()
	os.Exit(0)
}

func writeToFile(filename string, contents []byte) {
	err := ioutil.WriteFile(filename, contents, 0644)
	if err != nil {
		errorAndExit("Could not save gist content to file(s)!")
	}
}

func getFileContents(filename string) string {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		errorAndExit(err.Error())
	}
	return string(dat)
}

func loadSettings(file string) (map[string]string, error) {

	usr, _ := user.Current()
	dir := usr.HomeDir

	if file[:2] == "~/" {
		file = strings.Replace(file, "~/", dir+"/", 1)
	}

	contents, err := ioutil.ReadFile(file)
	conf := map[string]string{}
	if err == nil {
		lines := strings.Split(string(contents), "\n")
		for _, l := range lines {
			if strings.TrimSpace(l) == "" || string(strings.TrimSpace(l)[0:0]) == "#" {
				continue
			}
			parts := strings.Split(l, "=")
			conf[strings.Trim(parts[0], " ")] = strings.Trim(parts[1], " ")
		}
		return conf, nil
	}
	return conf, err
}

func errorAndExit(msg string) {
	fmt.Println("Error:", msg)
	fmt.Println()
	os.Exit(1)
}

func msgAndExit(msg string) {
	fmt.Println(msg)
	fmt.Println()
	os.Exit(0)
}

func centerString(str string, total_field_width int) string {

	str_len := len(str)
	spaces_to_pad := total_field_width - str_len
	if spaces_to_pad < 0 {
		spaces_to_pad = 0
	}
	var tmp_spaces float64
	var lr_spaces int

	tmp_spaces = float64(spaces_to_pad) / 2
	lr_spaces = int(tmp_spaces)

	buffer := bytes.NewBufferString("")

	spaces_remaining := total_field_width

	for i := 0; i < lr_spaces; i++ {
		buffer.WriteString(" ")
		spaces_remaining = spaces_remaining - 1
	}

	if spaces_to_pad == 0 {
		buffer.WriteString(str[:(total_field_width-4)] + "...")
	} else {
		buffer.WriteString(str)
	}

	spaces_remaining = spaces_remaining - str_len
	for i := spaces_remaining; i > 0; i-- {
		buffer.WriteString(" ")
	}

	return buffer.String()

}
