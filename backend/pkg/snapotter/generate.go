package snapotter

//go:generate sh -c "awk '/^ *operationId: generatePreview$/ {c++; if (c == 2) sub(/generatePreview$/, \"generatePreviewUpload\")} 1' openapi.yaml > snapotter_fixed.yaml && mv snapotter_fixed.yaml openapi.yaml"
//go:generate ogen --config .ogen.yml --clean openapi.yaml
