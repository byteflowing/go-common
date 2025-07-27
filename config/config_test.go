package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/byteflowing/go-common/crypto"
	"github.com/stretchr/testify/assert"
)

type ConfigTest struct {
	Int           int
	String        string
	IntDefault    int    `default:"1"`
	StringDefault string `default:"string_default"`
	SubConfig     ConfigTest2
	SubConfigPtr  *ConfigTest2
	SubConfigs    []ConfigTest2
	SubConfigsPtr []*ConfigTest2
	Map           map[string]*ConfigTest2
}

type ConfigTest2 struct {
	Int           int
	String        string
	IntDefault    int    `default:"1"`
	StringDefault string `default:"string_default"`
}

const jsonText string = `{
	"Int": 1,
	"String": "string",
	"IntDefault": 1,
	"StringDefault": "string_default",
	"SubConfig": {
		"Int": 1,
		"String": "string"
	},
	"SubConfigPtr": {
		"Int": 1,
		"String": "string"
	},
	"SubConfigs": [
		{
			"Int": 1,
			"String": "string1"
		}
	],
    "SubConfigsPtr": [
        {
            "Int": 1,
			"String": "string"
        }
    ],
	"Map": {
		"1": {
			"Int": 1,
			"String": "string"
		}
	}
}`

func TestReadConfig(t *testing.T) {
	tmpfile, err := createTempFile(".json", jsonText)
	assert.Nil(t, err)
	defer os.Remove(tmpfile)
	t1 := &ConfigTest{}
	type args struct {
		file   string
		config interface{}
	}
	tests := []struct {
		name    string
		args    args
		wants   ConfigTest
		wantErr bool
	}{
		{
			name: "test json file",
			args: args{file: tmpfile, config: t1},
			wants: ConfigTest{
				Int:           1,
				String:        "string",
				IntDefault:    1,
				StringDefault: "string_default",
				SubConfig: ConfigTest2{
					Int:           1,
					String:        "string",
					IntDefault:    1,
					StringDefault: "string_default",
				},
				SubConfigPtr: &ConfigTest2{
					Int:           1,
					String:        "string",
					IntDefault:    1,
					StringDefault: "string_default",
				},
				SubConfigs: []ConfigTest2{
					{
						Int:           1,
						String:        "string",
						IntDefault:    1,
						StringDefault: "string_default",
					},
				},
				SubConfigsPtr: []*ConfigTest2{
					&ConfigTest2{
						Int:           1,
						String:        "string",
						IntDefault:    1,
						StringDefault: "string_default",
					},
				},
				Map: map[string]*ConfigTest2{
					"1": {
						Int:           1,
						String:        "string",
						IntDefault:    1,
						StringDefault: "string_default",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.args.file)
			if err := ReadConfig(tt.args.file, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.wants.Int, t1.Int)
				assert.Equal(t, tt.wants.String, t1.String)
				assert.Equal(t, tt.wants.IntDefault, t1.IntDefault)
				assert.Equal(t, tt.wants.StringDefault, t1.StringDefault)

				assert.Equal(t, tt.wants.SubConfig.IntDefault, t1.SubConfig.IntDefault)
				assert.Equal(t, tt.wants.SubConfig.StringDefault, t1.SubConfig.StringDefault)

				assert.Equal(t, tt.wants.SubConfigPtr.IntDefault, t1.SubConfigPtr.IntDefault)
				assert.Equal(t, tt.wants.SubConfigPtr.StringDefault, t1.SubConfigPtr.StringDefault)

				assert.Equal(t, tt.wants.SubConfigs[0].IntDefault, t1.SubConfigs[0].IntDefault)
				assert.Equal(t, tt.wants.SubConfigs[0].StringDefault, t1.SubConfigs[0].StringDefault)

				assert.Equal(t, tt.wants.SubConfigsPtr[0].IntDefault, t1.SubConfigsPtr[0].IntDefault)
				assert.Equal(t, tt.wants.SubConfigsPtr[0].StringDefault, t1.SubConfigsPtr[0].StringDefault)

				assert.Equal(t, tt.wants.Map["1"].IntDefault, t1.Map["1"].IntDefault)
				assert.Equal(t, tt.wants.Map["1"].StringDefault, t1.Map["1"].StringDefault)

				//assert.Nil(t, tt.wants.SubConfigsOpt)
			}
			t.Logf("%v", tt.args.config)
		})
	}
}

func createTempFile(ext, text string) (string, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), crypto.Md5Hex([]byte(text))+"*"+ext)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tmpFile.Name(), []byte(text), os.ModeTemporary); err != nil {
		return "", err
	}

	filename := tmpFile.Name()
	if err = tmpFile.Close(); err != nil {
		return "", err
	}

	return filename, nil
}
