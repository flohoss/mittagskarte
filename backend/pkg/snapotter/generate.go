package snapotter

//go:generate sh -c "awk '/operationId: generatePreview/{c++} c==2{sub(/operationId: generatePreview/, \"operationId: generatePreviewUpload\")} 1' openapi.yaml > snapotter_fixed.yaml && mv snapotter_fixed.yaml openapi.yaml"
//go:generate ogen --config .ogen.yml --clean openapi.yaml
