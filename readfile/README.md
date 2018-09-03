# Read a File

Read an Image and convert the image to base 64 encoded string

## Installation

```bash
flogo install github.com/anshu2185/flogo-demo/readfile
```
Link for flogo web:
```
https://github.com/anshu2185/flogo-demo/readfile
```

## Schema
Inputs and Outputs:

```json
{
"inputs": [
        {
            "name": "filename",
            "type": "string",
            "required": true
        }
    ],
    "outputs": [
        {
            "name": "result",
            "type": "string"
        }
    ]
}
```
## Inputs
| Input    | Description                                                                 |
|:---------|:----------------------------------------------------------------------------|
| filename | The name of the Image you want to read (like `data.png` or `./tmp/data.png`) |

## Ouputs
| Output      | Description             |
|:------------|:------------------------|
| result      | The Base64 encoded content of the Image |
