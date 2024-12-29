package cmdtest

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var idExtractionRegex *regexp.Regexp = regexp.MustCompile(`\w+-(\d+)`)

func unmarshalAll[T any](body io.Reader, prefix string) (all []T, err error) {
	doc, err := html.Parse(body)

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

	namesToValues := fieldValuesFromDataNodes(node)
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

	return nil
}

func unmarshalUnassignedCount(bodyBytes io.Reader) (result UnassignedCount, err error) {

	rootNode, err := html.Parse(bodyBytes)
	if err != nil {
		return result, err
	}

	unassignedDivs := findEntityDivs(rootNode, "not-assigned")

	if len(unassignedDivs) > 1 {
		return result, errors.New("More than 1 unassigned entry found")
	}

	if len(unassignedDivs) == 0 {
		return result, nil
	}

	unassignedDiv := unassignedDivs[0]
	dataTags := findAllDataTags(unassignedDiv)

	if len(dataTags) != 1 {
		return result, fmt.Errorf("Expected exactly one data tag in unassigned entry but got: #%d", len(dataTags))
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
			return id, fmt.Errorf("Expected exactly two matches but got: %v", matches)
		}

		id = matches[1]
	}

	if id == "" {
		return id, errors.New("No id found")
	}

	return id, nil
}

func fieldValuesFromDataNodes(node *html.Node) map[string]string {
	dataNodes := findAllDataTags(node)

	namesToValues := make(map[string]string)

	for _, dataNode := range dataNodes {
		namesToValues[getClass(dataNode)] = getInnerTextData(dataNode)
	}

	return namesToValues
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
