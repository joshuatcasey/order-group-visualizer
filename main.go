package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/pack/pkg/dist"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide the location of a buildpack.toml file\n")
	}

	buildpackYaml := os.Args[1]
	fmt.Printf("Will look in file %s\n", buildpackYaml)

	buildpackDescriptor := dist.BuildpackDescriptor{}

	_, err := toml.DecodeFile(buildpackYaml, &buildpackDescriptor)
	if err != nil {
		log.Fatalf("Could not decode file %s\n", buildpackYaml)
	}

	buildpackIds := toNestedArray(buildpackDescriptor)

	var maxColumnSizes []int
	maxColumnSizes = findMaxColumnSizes(buildpackIds)

	for i, orderGroup := range buildpackIds {
		fmt.Printf("Order Group %d:", i+1)

		for j, id := range orderGroup {
			fmt.Printf(" %-*s", maxColumnSizes[j], id)
		}
		fmt.Printf("\n")
	}

	commonBeginningBuildpacks := findCommonBeginningElements(buildpackIds)
	commonEndingBuildpacks := findCommonEndingElements(buildpackIds)

	fmt.Print("Did we find any common beginning buildpacks?\n")
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

func toNestedArray(buildpackDescriptor dist.BuildpackDescriptor) [][]string {
	var result [][]string

	for _, orderGroup := range buildpackDescriptor.Order {
		var ids []string

		for _, buildpack := range orderGroup.Group {
			ids = append(ids, strings.TrimPrefix(buildpack.ID, "paketo-buildpacks/"))
		}

		result = append(result, ids)
	}

	return result
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
