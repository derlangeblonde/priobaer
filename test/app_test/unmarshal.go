package apptest

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"softbaer.dev/ass/internal/ui"
)

var idExtractionRegex = regexp.MustCompile(`\w+-(\d+)`)

func unmarshalAll[T any](body io.Reader, prefix string) (all []T, err error) {
	b, err := io.ReadAll(body)
	str := string(b)
	if err != nil {
		return all, err
	}

	doc, err := html.Parse(strings.NewReader(str))

	if err != nil {
		return all, err
	}

	divs := findEntityDivs(doc, prefix)

	for _, div := range divs {
		var single T

		err := unmarshal(&single, div)

		if err != nil {
			return all, err
		}

		all = append(all, single)
	}

	return all, nil
}

func unmarshal[T any](instance *T, node *html.Node) error {
	id, err := extractId(node)

	if err != nil {
		return err
	}

	v := reflect.ValueOf(instance).Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct pointer, got %T", instance)
	}

	namesToValues, namesToSliceValues := fieldValuesFromDataNodes(node)
	namesToValues["id"] = id

	typeOfT := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typeOfT.Field(i).Name
		fieldNameLower := strings.ToLower(fieldName)

		value, ok := namesToValues[fieldNameLower]

		if !ok {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int:
			intValue, err := strconv.Atoi(value)

			if err != nil {
				return err
			}

			field.SetInt(int64(intValue))
		}
	}

	participant, ok := any(instance).(*ui.Participant)

	if !ok {
		return nil
	}

	prioritizedCourseNames, ok := namesToSliceValues["priorities"]

	if !ok {
		return nil
	}
	for i, courseName := range prioritizedCourseNames {
		level := uint8(i + 1)
		participant.Priorities = append(participant.Priorities, ui.Priority{CourseName: courseName, Level: level})
	}

	return nil
}

func unmarshalUnassignedCount(bodyBytes io.Reader) (result UnassignedCount, err error) {

	rootNode, err := html.Parse(bodyBytes)
	if err != nil {
		return result, err
	}

	unassignedDivs := findEntityDivs(rootNode, "not-assigned")

	if len(unassignedDivs) > 1 {
		return result, errors.New("more than 1 unassigned entry found")
	}

	if len(unassignedDivs) == 0 {
		return result, nil
	}

	unassignedDiv := unassignedDivs[0]
	dataTags := findAllDataTags(unassignedDiv)

	if len(dataTags) != 1 {
		return result, fmt.Errorf("expected exactly one data tag in unassigned entry but got: #%d", len(dataTags))
	}

	dataTag := dataTags[0]

	count, err := strconv.Atoi(getInnerTextData(dataTag))

	if err != nil {
		return result, err
	}

	result = UnassignedCount{Value: count, Updated: true}

	return result, nil
}

func extractId(node *html.Node) (id string, err error) {
	for _, attr := range node.Attr {
		if attr.Key != "id" {
			continue
		}

		matches := idExtractionRegex.FindStringSubmatch(attr.Val)

		if len(matches) != 2 {
			return id, fmt.Errorf("expected exactly two matches but got: %v", matches)
		}

		id = matches[1]
	}

	if id == "" {
		return id, errors.New("no id found")
	}

	return id, nil
}

var slicePropertyNameRegex = regexp.MustCompile(`^(\w+)-(\d+)$`)

func fieldValuesFromDataNodes(node *html.Node) (map[string]string, map[string][]string) {
	dataNodes := findAllDataTags(node)

	namesToValues := make(map[string]string)
	namesToSliceValues := make(map[string][]string)

	for _, dataNode := range dataNodes {
		name := getClass(dataNode)
		match := slicePropertyNameRegex.FindStringSubmatch(name)

		if len(match) == 3 {
			name = match[1]
			// TODO: not sure if index is important or whether we can rely on
			// that dataNodes come in the correct order?
			// index := match[2]
			existingSlice, ok := namesToSliceValues[name]
			if ok {
				namesToSliceValues[name] = append(existingSlice, getInnerTextData(dataNode))
			} else {
				namesToSliceValues[name] = []string{getInnerTextData(dataNode)}
			}
		} else {
			namesToValues[name] = getInnerTextData(dataNode)
		}

	}

	return namesToValues, namesToSliceValues
}

func findAllDataTags(node *html.Node) []*html.Node {
	if node.Type == html.ElementNode && node.Data == "data" && node.FirstChild != nil {
		return []*html.Node{node}
	}

	nodes := make([]*html.Node, 0)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, findAllDataTags(c)...)
	}

	return nodes
}

func getClass(node *html.Node) string {

	for _, attribute := range node.Attr {
		if attribute.Key == "class" {
			return attribute.Val
		}
	}

	panic("getClass invoked on node that has no class")
}

func getInnerTextData(node *html.Node) string {
	inner := node.FirstChild

	if inner.Type != html.TextNode {
		panic("getInnerTextData called on node which FirstChild is not a TextNode")
	}

	return strings.TrimSpace(inner.Data)
}

func findEntityDivs(current *html.Node, idContains string) []*html.Node {
	if current.Type == html.ElementNode && current.Data == "div" {
		for _, attr := range current.Attr {
			if attr.Key == "id" && strings.Contains(attr.Val, idContains) {
				return []*html.Node{current}
			}
		}
	}

	alreadyFound := make([]*html.Node, 0)
	for c := current.FirstChild; c != nil; c = c.NextSibling {
		alreadyFound = append(alreadyFound, findEntityDivs(c, idContains)...)
	}

	return alreadyFound
}
