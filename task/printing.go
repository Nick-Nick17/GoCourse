//go:build !solution

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"golang.org/x/exp/maps"
)

func printingTab(authors []*Author) {
	const padding = 1
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Name\tLines\tCommits\tFiles\t")
	for _, author := range authors {
		fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", author.Name, author.Lines, author.Commits, author.Files)
	}
	w.Flush()
}

func printingCSV(authors []*Author) {
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"Name", "Lines", "Commits", "Files"})
	for _, author := range authors {
		w.Write([]string{author.Name, fmt.Sprint(author.Lines), fmt.Sprint(author.Commits), fmt.Sprint(author.Files)})
	}
	w.Flush()
}

func printingJSON(authors []*Author) {
	js, _ := json.Marshal(authors)
	fmt.Println(string(js))
}

func printingJSONLines(authors []*Author) {
	for _, author := range authors {
		js, _ := json.Marshal(author)
		fmt.Println(string(js))
	}
}

func printResults(authors map[string]*Author) {
	//  забрать всех авторов, отсортировать
	res := maps.Values(authors)
	switch options.orderBy {
	case "lines":
		// default
		// (lines, commits, files)
		sort.Slice(res, func(i, j int) bool {
			return res[i].Lines > res[j].Lines ||
				res[i].Lines == res[j].Lines && (res[i].Commits > res[j].Commits ||
					res[i].Commits == res[j].Commits && (res[i].Files > res[j].Files ||
						res[i].Files == res[j].Files && res[i].Name < res[j].Name))
		})
	case "commits":
		// (commits, lines, files)
		sort.Slice(res, func(i, j int) bool {
			return res[i].Commits > res[j].Commits ||
				res[i].Commits == res[j].Commits && (res[i].Lines > res[j].Lines ||
					res[i].Lines == res[j].Lines && (res[i].Files > res[j].Files ||
						res[i].Files == res[j].Files && res[i].Name < res[j].Name))
		})
	case "files":
		// (files, lines, commits)
		sort.Slice(res, func(i, j int) bool {
			return res[i].Files > res[j].Files ||
				res[i].Files == res[j].Files && (res[i].Lines > res[j].Lines ||
					res[i].Lines == res[j].Lines && (res[i].Commits > res[j].Commits ||
						res[i].Commits == res[j].Commits && res[i].Name < res[j].Name))
		})
	}

	switch options.format {
	case "tabular":
		printingTab(res)
	case "csv":
		printingCSV(res)
	case "json":
		printingJSON(res)
	case "json-lines":
		printingJSONLines(res)
	}
}
