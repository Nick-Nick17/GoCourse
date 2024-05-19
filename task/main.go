//go:build !solution

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	pflag "github.com/spf13/pflag"
)

type Author struct {
	Name    string
	Lines   int
	Commits int
	Files   int
}

var options struct {
	// Расчёт
	repository   string
	revision     string
	useCommitter bool

	// Вывод
	orderBy string
	format  string

	// Ограничения на файл
	extensions *[]string
	languages  *[]string
	exclude    *[]string
	restrictTo *[]string

	// Обработанные настройки
}

type Data struct {
	mutex   sync.Mutex
	authors map[string]*Author
}

func NewData() *Data {
	return &Data{
		authors: make(map[string]*Author),
	}
}

// Improve
func gitBlame(data *Data, file string) {
	cmd := exec.Command("git", "blame", "--incremental", options.revision, "--", file)
	//cmd.Dir = options.repository
	output, err := cmd.Output()
	if err != nil {
		return
	}

	if string(output) == "" {
		rawData, err := exec.Command("git", "log", "--pretty=format:\"%H$^!~&%an$^!~&%cn\"", options.revision, "--", file).Output()
		if err != nil {
			fmt.Println(err)
			return
		}
		parsed := strings.Split(string(rawData), "\n")[0]
		parsed = parsed[1 : len(parsed)-1]
		info := strings.Split(parsed, "$^!~&")
		author := info[1] // author
		if options.useCommitter {
			author = info[2] // committer
		}
		if _, ok := data.authors[author]; !ok {
			data.authors[author] = &Author{Name: author, Files: 1, Commits: 1}
		} else {
			data.authors[author].Files++
			data.authors[author].Commits++
		}
		return
	}

	lines := bytes.Split(output, []byte("\n"))
	authors := make(map[string]*Author)

	var last_author string
	for index, line := range lines {
		if len(string(line)) == 0 {
			continue
		}
		sha1Pattern := regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
		words := strings.Split(string(line), " ")
		if !sha1Pattern.MatchString(words[0]) {
			continue
		}

		if len(lines) <= index+1 {
			continue
		}

		updateCommits := false
		var name string

		if !options.useCommitter {
			if len(string(lines[index+1])) == 0 {
				continue
			}
			smb := string(lines[index+1])[0]
			if smb != 'a' && smb != 'p' {
				continue
			}

			if smb == 'p' {
				name = last_author
			} else {
				updateCommits = true
				name = string(lines[index+1])[len("author "):]
			}
		}

		if options.useCommitter {
			if len(lines) <= index+5 {
				continue
			}
			if len(string(lines[index+5])) == 0 {
				continue
			}

			smb1 := string(lines[index+1])[0]
			smb2 := string(lines[index+5])[0]
			if smb1 != 'p' && smb2 != 'c' {
				continue
			}

			if smb1 == 'p' {
				name = last_author
			} else {
				updateCommits = true
				name = string(lines[index+5])[len("committer "):]
			}
		}

		if _, ok := authors[name]; !ok {
			authors[name] = &Author{Name: name, Files: 1}
		}
		lns, err := strconv.Atoi(words[len(words)-1]) // Improve
		if err != nil {
			continue
		}
		authors[name].Lines += lns
		if updateCommits {
			authors[name].Commits++
		}
		last_author = name
	}

	for _, author := range authors {
		if _, ok := data.authors[author.Name]; !ok {
			data.authors[author.Name] = author
		} else {
			data.authors[author.Name].Lines += author.Lines
			data.authors[author.Name].Files += author.Files
			data.authors[author.Name].Commits += author.Commits
		}
	}
}

func GetFiles() []string {
	// Находим файлы
	err := os.Chdir(options.repository)
	if err != nil {
		fmt.Println("Error getting repository files:", err)
		os.Exit(1)
	}
	filesCmd := exec.Command("git", "ls-tree", "-r", "--name-only", options.revision)
	filesCmd.Dir = options.repository
	filesOutput, err := filesCmd.Output()
	if err != nil {
		fmt.Println("Error getting repository files:", err)
		os.Exit(1)
	}

	filesList := strings.Split(string(filesOutput), "\n")
	return filesList
}

func main() {
	// Обработка флагов
	pflag.StringVar(&options.repository, "repository", ".", "путь до Git репозитория; по умолчанию текущая директория")
	pflag.StringVar(&options.revision, "revision", "HEAD", "указатель на коммит; HEAD по умолчанию")
	pflag.StringVar(&options.orderBy, "order-by", "lines", "ключ сортировки результатов; один из `lines` (дефолт), `commits`, `files`")
	pflag.StringVar(&options.format, "format", "tabular", "формат вывода; один из `tabular` (дефолт), `csv`, `json`, `json-lines`")

	options.extensions = pflag.StringSlice("extensions", nil, "список расширений, сужающий список файлов в расчёте; множество ограничений разделяется запятыми, например, '.go,.md'") // возможные окончания файла
	options.languages = pflag.StringSlice("languages", nil, "список языков (программирования, разметки и др.), сужающий список файлов в расчёте; множество ограничений разделяется запятыми, например `'go,markdown'`")
	options.exclude = pflag.StringSlice("exclude", nil, "набор [Glob](https://en.wikipedia.org/wiki/Glob_(programming)) паттернов, исключающих файлы из расчёта, например `'foo/*,bar/*'`")
	options.restrictTo = pflag.StringSlice("restrict-to", nil, "набор Glob паттернов, исключающий все файлы, не удовлетворяющие ни одному из паттернов набора")

	pflag.BoolVar(&options.useCommitter, "use-committer", false, "булев флаг, заменяющий в расчётах автора (дефолт) на коммиттера")
	pflag.Parse()

	data := NewData()
	filesList := GetFiles()

	var wg sync.WaitGroup
	for _, file := range filesList {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			// Предобработка, проверка на паттерны
			if !isFileSuitable(file) {
				return
			}

			// Critical section: working with map
			data.mutex.Lock()
			gitBlame(data, file)
			data.mutex.Unlock()
		}(file)
	}
	wg.Wait()

	printResults(data.authors)
}
