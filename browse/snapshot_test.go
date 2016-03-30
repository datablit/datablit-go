package browse

import (
	"testing"
	"github.com/blitter/meta/yang"
	"github.com/blitter/node"
	"strings"
	"bytes"
	"fmt"
)

func TestSnapshotRestore(t *testing.T) {

	tests := []struct {
		snapshot string
		expected string
	}{
		{
			snapshot : `
{
  "meta": {
    "container": {
      "ident": "hobbies",
      "definitions": [
        {
          "ident": "birding",
          "container": {
            "ident": "birding",
            "definitions": [
              {
                "ident": "favorite-species",
                "leaf": {
                  "ident": "favorite-species",
                  "type": {
                    "ident": "string"
                  }
                }
              }
            ]
          }
        }
      ]
    }
  },
  "data": {
    "hobbies": {
      "birding": {
        "favorite-species": "towhee"
      }
    }
  }
}`,
			expected : `{"birding":{"favorite-species":"towhee"}}`,
		},
		{
			snapshot: `
{
  "meta": {
    "list": {
      "ident": "hobbies",
      "definitions": [
        {
          "ident": "name",
          "leaf": {
            "ident": "name",
            "type": {
              "ident": "string"
            }
          }
        },
        {
          "ident": "favorite",
          "container": {
            "ident": "favorite",
            "definitions": [
              {
                "ident": "label",
                "leaf": {
                  "ident": "label",
                  "type": {
                    "ident": "string"
                  }
                }
              }
            ]
          }
        }
      ]
    }
  },
  "data": {
    "hobbies": [
      {
        "name": "birding",
        "favorite": {
          "label": "towhee"
        }
      }
    ]
  }
}`,
			expected: `{"hobbies":[{"name":"birding","favorite":{"label":"towhee"}}]}`,
		},
		{
			snapshot: `
{
  "meta": {
    "list-item": {
      "ident": "hobbies",
      "key" : ["name"],
      "definitions": [
        {
          "ident": "name",
          "leaf": {
            "ident": "name",
            "type": {
              "ident": "string"
            }
          }
        },
        {
          "ident": "favorite",
          "container": {
            "ident": "favorite",
            "definitions": [
              {
                "ident": "label",
                "leaf": {
                  "ident": "label",
                  "type": {
                    "ident": "string"
                  }
                }
              }
            ]
          }
        }
      ]
    },
    "key": [
      "birding"
    ]
  },
  "data": {
    "hobbies": [
      {
        "name": "birding",
        "favorite": {
          "label": "towhee"
        }
      }
    ]
  }
}			`,
			expected: `{"hobbies":[{"name":"birding","favorite":{"label":"towhee"}}]}`,
		},
	}

	c := node.NewContext()
	for i, test := range tests {
		in := node.NewJsonReader(strings.NewReader(test.snapshot)).Node()
		snap, err := RestoreSelection(in)
		if err != nil {
			t.Errorf("#%d - %s", i, err.Error())
			continue
		}
		var actualBytes bytes.Buffer
		if err = c.Selector(snap).InsertInto(node.NewJsonWriter(&actualBytes).Node()).LastErr; err != nil {
			t.Errorf("#%d - %s", i, err.Error())
			continue
		}
		actual := actualBytes.String()
		if actual != test.expected {
			t.Errorf("#%d - %s", i, actual)
		}
	}
}

func TestSnapshotSave(t *testing.T) {
	moduleStr := `
module test {
	prefix "t";
	namespace "t";
	revision 0;

        %s

	container hockey {
		leaf favorite-team {
			type string;
		}
	}
}`
	tests := []struct {
		yang string
		data string
		url string
		expected string
		roundtrip string
	}{
		{
			yang :
				`
					container hobbies {
						container birding {
							leaf favorite-species {
								type string;
							}
						}
					}
				`,
			data :
				`{
					"hobbies" : {
						"birding" : {
							"favorite-species" : "towhee"
						}
					}
				}`,
			url : "hobbies",
			expected :
				`"data":{"hobbies":{"birding":{"favorite-species":"towhee"}}}`,
			roundtrip :
				`{"birding":{"favorite-species":"towhee"}}`,

		},
		{
			yang :
			`
				list hobbies {
					key "name";
					leaf name {
						type string;
					}
					container favorite {
						leaf label {
							type string;
						}
					}
				}
			`,
			data :
			`{
				"hobbies" : [{
					"name" : "birding",
					"favorite" : {
						"label" : "towhee"
					}
				}]
			}`,
			url : "hobbies",
			expected :
				`"data":{"hobbies":[{"name":"birding","favorite":{"label":"towhee"}}]}}`,
			roundtrip:
				`{"hobbies":[{"name":"birding","favorite":{"label":"towhee"}}]}`,
		},
		{
			yang :
			`
				list hobbies {
					key "name";
					leaf name {
						type string;
					}
					container favorite {
						leaf label {
							type string;
						}
					}
				}
			`,
			data :
			`{
				"hobbies" : [{
					"name" : "birding",
					"favorite" : {
						"label" : "towhee"
					}
				}]
			}`,
			url : "hobbies=birding",
			expected :
				`"data":{"hobbies":[{"name":"birding","favorite":{"label":"towhee"}}]}}`,
			roundtrip:
				`{"hobbies":[{"name":"birding","favorite":{"label":"towhee"}}]}`,
		},
	}
	for i, test := range tests {
		mstr := fmt.Sprintf(moduleStr, test.yang)
		mod, err := yang.LoadModuleCustomImport(mstr, nil)
		if err != nil {
			panic(err)
		}
		n := node.NewJsonReader(strings.NewReader(test.data)).Node()
		c := node.NewContext()
		sel := c.Select(mod, n).Find(test.url)
		if sel.LastErr != nil {
			t.Error("#%d - %s", i, sel.LastErr.Error())
			continue
		}
		snap := SaveSelection(sel.Selection)
		var actualBytes bytes.Buffer
		if err = c.Selector(snap).InsertInto(node.NewJsonWriter(&actualBytes).Node()).LastErr; err != nil {
			t.Errorf("#%d - %s", i, err.Error())
			continue
		}
		actual := actualBytes.String()
		if !strings.Contains(actual, test.expected) {
			t.Errorf("#%d - %s", i, actual)
			continue
		}

		roundtrip, rtErr := RestoreSelection(node.NewJsonReader(&actualBytes).Node())
		if rtErr != nil {
			t.Errorf("#%d roundtrip - %s", i, rtErr.Error())
			continue
		}
		var roundtripBytes bytes.Buffer
		if restoreErr := c.Selector(roundtrip).InsertInto(node.NewJsonWriter(&roundtripBytes).Node()).LastErr; restoreErr != nil {
			t.Errorf("#%d roundtrip restore - %s", i, restoreErr.Error())
			continue
		}
		roundtripActual := roundtripBytes.String()
		if roundtripActual != test.roundtrip {
			t.Errorf("#%d roundtrip wrong expectation. actual:%s", i, roundtripActual)
			continue
		}
	}
}
