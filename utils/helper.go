package utils

import (
	"fmt"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xpath"
	"net/url"
	"sort"
	"strings"
)

func PrepareFields(fields map[string]string) string {
	var keyFields []string
	for key, _ := range fields {
		keyFields = append(keyFields, key)
	}
	sort.Strings(keyFields)

	var dataFields []string
	for _, key := range keyFields {
		dataFields = append(dataFields, key+"="+fields[key])
	}

	return strings.Join(dataFields, "&")
}

func EncodeUrl(host string, path string, fields map[string]string) (string, error) {
	if urlBuilder, err := url.Parse(host); err != nil {
		return "", err
	} else {
		urlBuilder.Path += path

		parameters := url.Values{}
		for k, v := range fields {
			parameters.Add(k, v)
		}
		urlBuilder.RawQuery = parameters.Encode()

		return urlBuilder.String(), nil
	}
}

func DecodeUrl(link string) (string, string, error) {
	if du, err := url.Parse(link); err != nil {
		return "", "", err
	} else {
		return du.Host, du.Path, nil
	}
}

func GetContent(html string, path string) (string, error) {
	doc, _ := gokogiri.ParseHtml([]byte(html))
	defer doc.Free()

	if nodes, err := doc.Root().Search(xpath.Compile(path)); err != nil {
		return "", fmt.Errorf("GetContent helper: %v\n", err.Error())
	} else {
		for _, node := range nodes {
			return node.Content(), nil
		}

		return "", fmt.Errorf("GetContent helper: No results for XPATH %v\n", path)
	}
}

func GetStatusFromNodeCss(html string, path string) (string, string, error) {
	doc, _ := gokogiri.ParseHtml([]byte(html))
	defer doc.Free()

	if nodes, err := doc.Root().Search(xpath.Compile(path)); err != nil {
		return "", "", fmt.Errorf("GetStatusFromNodeCss helper: %v\n", err.Error())
	} else {
		for _, node := range nodes {
			if id := node.Attribute("id"); id == nil {
				return "", "", fmt.Errorf("GetStatusFromNodeCss helper: Attribute \"id\" not found\n")
			} else if class := node.Attribute("class"); class == nil {
				return "", "", fmt.Errorf("GetStatusFromNodeCss helper: Attribute \"class\" not found\n")
			} else {
				switch class.Content() {
				case "enabled":
					return id.Content(), "Connected", nil

				case "disabled":
					return id.Content(), "Not connected", nil

				case "unused":
					return id.Content(), "Unused", nil

				default:
					return "", "", fmt.Errorf("GetStatusFromNodeCss helper: Unsupported case %v %v\n", id.Content(), class.Content())
				}
			}
		}

		return "", "", fmt.Errorf("GetStatusFromNodeCss helper: No results for XPATH %v\n", path)
	}
}

func GetTableFromArray(html string, path string) ([]string, []string, error) {
	doc, _ := gokogiri.ParseHtml([]byte(html))
	defer doc.Free()

	if nodes, err := doc.Root().Search(xpath.Compile(path)); err != nil {
		return nil, nil, fmt.Errorf("GetTableFromArray helper: %v\n", err.Error())
	} else if len(nodes) == 0 {
		return nil, nil, fmt.Errorf("GetTableFromArray helper: No results for XPATH %v\n", path)
	} else {
		var labels, values []string

		for _, node := range nodes {
			if childNodes, err := node.Search(xpath.Compile(".//th")); err != nil {
				return nil, nil, fmt.Errorf("GetTableFromArray helper: %v\n", err.Error())
			} else {
				for _, childNode := range childNodes {
					labels = append(labels, strings.TrimSpace(childNode.Content()))
				}

				if childNodes, err := node.Search(xpath.Compile(".//td")); err != nil {
					return nil, nil, fmt.Errorf("GetTableFromArray helper: %v\n", err.Error())
				} else {
					for _, childNode := range childNodes {
						iVal := strings.TrimSpace(childNode.Content())
						if strings.HasPrefix(iVal, ":") {
							iVal = strings.Split(iVal, " ")[1]
						}
						values = append(values, removePunctuation(iVal))
					}
				}

				if len(labels) != len(values) {
					return nil, nil, fmt.Errorf("GetTableFromArray helper: Number of labels and values differ\n")
				}
			}
		}

		return labels, values, nil
	}
}

func GetTableDataAsArrayWithHeaders(html string, path string, excludeRows int, excludeCols int) ([]string, [][]string, error) {
	doc, _ := gokogiri.ParseHtml([]byte(html))
	defer doc.Free()

	if doc.Root() == nil {
		return nil, nil, nil
	} else if cols, err := doc.Root().Search(xpath.Compile(path + "/thead/tr/th")); err != nil {
		return nil, nil, fmt.Errorf("GetTableDataAsArrayWithHeaders helper: %v\n", err.Error())
	} else if len(cols) == 0 {
		return nil, nil, fmt.Errorf("GetTableDataAsArrayWithHeaders helper: No results for XPATH %v\n", path)
	} else {
		var labels []string

		for _, col := range cols[:len(cols)-excludeCols] {
			labels = append(labels, removePunctuation(col.Content()))
		}

		if rows, err := doc.Root().Search(xpath.Compile(path + "/tbody/tr")); err != nil {
			return nil, nil, fmt.Errorf("GetTableDataAsArrayWithHeaders helper: %v\n", err.Error())
		} else {
			nbRows := len(rows) - excludeRows
			values := make([][]string, nbRows)
			for i, _ := range rows[:nbRows] {
				cols, _ = rows[i].Search(".//td")
				for _, col := range cols[:len(cols)-excludeCols] {
					if hidden_vals, err := col.Search(".//a"); err != nil || hidden_vals == nil {
						values[i] = append(values[i], removePunctuation(col.Content()))
					} else {
						for _, hidden_val := range hidden_vals {
							var urls []string
							var str = strings.Trim(removePunctuation(hidden_val.Attribute("value").Content()), "{}")
							for _, j := range strings.Split(str, ",") {
								for _, k := range strings.Split(j, "\":\"") {
									if strings.HasPrefix(k, "\"") {
										urls = append(urls, strings.Replace(k, "\\/", "/", -1) + "\"")
									}
								}
							}
							values[i] = append(values[i], strings.Join(urls, ";"))
						}
					}
				}
			}

			return labels, values, nil
		}
	}
}

func GetFormValues(html string, path string) (map[string]string, error) {
	formData := make(map[string]string)

	doc, _ := gokogiri.ParseHtml([]byte(html))
	defer doc.Free()

	if inputNodes, err := doc.Root().Search(xpath.Compile(path + "//input" + " | " + path + "//select" + " |" + path + "//button")); err != nil {
		return nil, fmt.Errorf("GetFormValues helper: %v\n", err.Error())
	} else if len(inputNodes) == 0 {
		return nil, fmt.Errorf("GetFormValues helper: No results for XPATH %v\n", path)
	} else {
		for _, inputNode := range inputNodes {
			if inputNode.Name() == "input" {
				switch inputNode.Attribute("type").Content() {
				case "radio":
					if checked := inputNode.Attribute("checked"); checked != nil && checked.Content() == "checked" {
						formData[inputNode.Attribute("name").Content()] = inputNode.Attribute("value").Content()
					}
					break

				default:
					if nameNode := inputNode.Attribute("name"); nameNode != nil {
						name := nameNode.Content()
						if formData[name] = ""; inputNode.Attribute("value") != nil {
							formData[name] = inputNode.Attribute("value").Content()
						}
					}
					break
				}
			} else if inputNode.Name() == "button" {
				if nameNode := inputNode.Attribute("name"); nameNode != nil {
					formData[nameNode.Content()] = nameNode.Content()
				}
			} else if inputNode.Name() == "select" {
				optionNodes, _ := inputNode.Search(xpath.Compile(".//option"))
				for _, optionNode := range optionNodes {
					if optionNode.Attribute("selected") != nil && optionNode.Attribute("selected").Content() == "selected" {
						formData[inputNode.Attribute("name").Content()] = optionNode.Attribute("value").Content()
					}
				}
			}
		}

		return formData, nil
	}
}
