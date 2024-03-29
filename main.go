package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/pack/pkg/dist"
	"golang.org/x/exp/slices"
)

func main() {
	var (
		buildpackTomlPath string
		outputFormat      string
		uniqueOnly        bool
		requiredOnly      bool
		withVersion       bool
	)

	flag.StringVar(&buildpackTomlPath, "buildpack-toml-path", "", "full path to a meta buildpack's buildpack.toml file (OPTIONAL, default='./buildpack.toml')")
	flag.StringVar(&outputFormat, "output-format", "", "output format (OPTIONAL, default='short') [table, short, short-json, hist]")
	flag.BoolVar(&uniqueOnly, "unique-only", false, "only print unique buildpack ids (OPTIONAL, default='false')")
	flag.BoolVar(&requiredOnly, "required-only", false, "only print required buildpack ids (OPTIONAL, default='false')")
	flag.BoolVar(&withVersion, "with-version", false, "print buildpack version as well as id (OPTIONAL, default='false')")
	flag.Parse()

	if buildpackTomlPath == "" {
		buildpackTomlPath = "./buildpack.toml"
	}

	if outputFormat == "" {
		outputFormat = "short"
	}

	switch outputFormat {
	case "short-json":
	default:
		fmt.Printf("Will look in file %s\n", buildpackTomlPath)
	}

	buildpackDescriptor := dist.BuildpackDescriptor{}

	_, err := toml.DecodeFile(buildpackTomlPath, &buildpackDescriptor)
	if err != nil {
		log.Fatalf("Could not decode file %s\n", buildpackTomlPath)
	}

	buildpackIds := toNestedArray(buildpackDescriptor, requiredOnly, uniqueOnly, withVersion)

	switch outputFormat {
	case "table":
		printTable(buildpackIds)
	case "short":
		printShortList(buildpackIds)
	case "short-json":
		printShortJsonList(buildpackIds)
	case "hist":
		printHistogram(buildpackIds)
	default:
		log.Fatalf("--output-format not recognized")
	}
}

func printHistogram(buildpackIds [][]string) {
	idToCount := make(map[string]int)

	for _, orderGroup := range buildpackIds {
		for _, id := range orderGroup {
			idToCount[id]++
		}
	}

	countToId := make([][]string, 0)

	for id, count := range idToCount {
		for len(countToId) <= count {
			countToId = append(countToId, make([]string, 0))
		}

		countToId[count] = append(countToId[count], id)
	}

	fmt.Printf("Histogram:\n")

	for i := len(countToId) - 1; i >= 0; i-- {
		if len(countToId[i]) > 0 {
			fmt.Printf("%d: %s\n", i, strings.Join(countToId[i], ", "))
		}
	}
}

func printShortList(buildpackIds [][]string) {
	commonBeginningBuildpacks := findCommonBeginningElements(buildpackIds)
	commonEndingBuildpacks := findCommonEndingElements(buildpackIds)

	if len(commonEndingBuildpacks) < 1 {
		fmt.Print("No common beginning buildpacks found\n")
	} else {
		fmt.Print("Common beginning buildpacks\n")
	}
	for i := 0; i < len(commonBeginningBuildpacks); i++ {
		fmt.Printf("- %s\n", commonBeginningBuildpacks[i])
	}

	for i, orderGroup := range buildpackIds {
		fmt.Printf("Order Group #%d, with %d buildpacks\n", i+1, len(orderGroup))

		for j, id := range orderGroup {
			if j < len(commonBeginningBuildpacks) {
				continue
			}
			leftToGo := len(orderGroup) - j
			if leftToGo <= len(commonEndingBuildpacks) && id == commonEndingBuildpacks[leftToGo-1] {
				continue
			}
			fmt.Printf("- (#%02d): %s\n", j+1, id)
		}
	}

	fmt.Print("Common ending buildpacks\n")
	for i := len(commonEndingBuildpacks) - 1; i >= 0; i-- {
		fmt.Printf("- %s\n", commonEndingBuildpacks[i])
	}
}

type ShortJson struct {
	Beginning   []string   `json:"beginning"`
	Ending      []string   `json:"ending"`
	OrderGroups [][]string `json:"order_groups"`
}

func printShortJsonList(buildpackIds [][]string) {
	shortJson := ShortJson{
		Beginning: findCommonBeginningElements(buildpackIds),
		Ending:    findCommonEndingElements(buildpackIds),
	}

	shortJson.OrderGroups = make([][]string, len(buildpackIds))

	for i, orderGroup := range buildpackIds {
		shortJson.OrderGroups[i] = make([]string, 0)
		for j, id := range orderGroup {
			if j < len(shortJson.Beginning) {
				continue
			}
			leftToGo := len(orderGroup) - j
			if leftToGo <= len(shortJson.Ending) && id == shortJson.Ending[leftToGo-1] {
				continue
			}
			shortJson.OrderGroups[i] = append(shortJson.OrderGroups[i], id)
		}
	}

	shortJsonString, err := json.MarshalIndent(shortJson, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Print(string(shortJsonString))
}

func printTable(buildpackIds [][]string) {
	var maxColumnSizes []int
	maxColumnSizes = findMaxColumnSizes(buildpackIds)

	for i, orderGroup := range buildpackIds {
		fmt.Printf("Order Group %d:", i+1)

		for j, id := range orderGroup {
			if j == 0 {
				fmt.Printf(" %-*s", maxColumnSizes[j], id)
			} else {
				fmt.Printf(" | %-*s", maxColumnSizes[j], id)
			}
		}
		fmt.Printf("\n")
	}
}

func findMaxColumnSizes(buildpackIds [][]string) []int {
	var result []int

	for _, orderGroup := range buildpackIds {
		for len(result) < len(orderGroup) {
			result = append(result, 0)
		}

		for j, id := range orderGroup {
			if len(id) > result[j] {
				result[j] = len(id)
			}
		}
	}

	return result
}

func toNestedArray(buildpackDescriptor dist.BuildpackDescriptor, requiredOnly, uniqueOnly, withVersion bool) [][]string {
	var result [][]string

	var alreadySeen []string

	for _, orderGroup := range buildpackDescriptor.Order {
		var ids []string

		for _, buildpack := range orderGroup.Group {
			if !requiredOnly || (requiredOnly && !buildpack.Optional) {
				id := toString(buildpack, withVersion)

				if !uniqueOnly || (uniqueOnly && !slices.Contains(alreadySeen, id)) {
					ids = append(ids, id)
					alreadySeen = append(alreadySeen, id)
				}
			}
		}

		result = append(result, ids)
	}

	return result
}

func toString(buildpack dist.BuildpackRef, withVersion bool) string {
	version := buildpack.Version

	if version == "" {
		version = "<UNKNOWN-VERSION>"
	}

	id := strings.TrimPrefix(buildpack.ID, "paketo-buildpacks/")

	if !withVersion {
		return id
	}

	return id + "@" + version
}

func findCommonBeginningElements(buildpackIds [][]string) []string {
	common := buildpackIds[0]
	countCommon := 0

	for i := 0; i < len(common); i++ {
		allMatch := true
		for j := 1; j < len(buildpackIds); j++ {
			orderGroup := buildpackIds[j]
			if len(orderGroup) < i {
				allMatch = false
				break
			}

			if orderGroup[i] != common[i] {
				allMatch = false
				break
			}
		}

		if allMatch {
			countCommon = i + 1
		} else {
			break
		}
	}

	if countCommon > 0 {
		return common[:countCommon]
	}

	return []string{}
}

func findCommonEndingElements(buildpackIds [][]string) []string {
	buildPackIdsReversed := make([][]string, len(buildpackIds))

	for i := 0; i < len(buildpackIds); i++ {
		buildPackIdsReversed[i] = swap(buildpackIds[i])
	}

	return findCommonBeginningElements(buildPackIdsReversed)
}

func swap(slice []string) []string {
	switch len(slice) {
	case 0:
		return slice
	case 1:
		return slice
	}
	result := make([]string, len(slice))
	half := len(slice) / 2

	for i := 0; i < half; i++ {
		result[i] = slice[len(slice)-i-1]
		result[len(slice)-i-1] = slice[i]
	}

	isOdd := len(slice)%2 == 1

	if isOdd {
		result[half] = slice[half]
	}

	return result
}
