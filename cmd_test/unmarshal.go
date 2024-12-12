package cmdtest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

func unmarshal[T any](instance *T, node *html.Node) error {
	v := reflect.ValueOf(instance).Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct pointer, got %T", instance)
	}

	namesToValues := fieldValuesFromDataNodes(node)

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

func fieldValuesFromDataNodes(node *html.Node) map[string]string {
	dataNodes := findAllDataTags(node)

	namesToValues := make(map[string]string)

	for _, dataNode := range dataNodes {
		namesToValues[getClass(dataNode)] = getInnerTextData(dataNode)
	}

	return namesToValues
}

func findAllDataTags(node *html.Node) []*html.Node {
	if node.Type == html.ElementNode && node.Data == "data" {
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

func findEntityDivs(current *html.Node, prefix string) []*html.Node {
	if current.Type == html.ElementNode && current.Data == "div" {
		for _, attr := range current.Attr {
			if attr.Key == "id" && strings.Contains(attr.Val, prefix) {
				return []*html.Node{current}
			}
		}
	}

	alreadyFound := make([]*html.Node, 0)
	for c := current.FirstChild; c != nil; c = c.NextSibling {
		alreadyFound = append(alreadyFound, findEntityDivs(c, prefix)...)
	}

	return alreadyFound
}
