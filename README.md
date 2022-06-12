# AI2api

## requirements
Because we add capability of image processing by import a library, we need to install libvips 8.3+;<br>
More pls refer [this](https://github.com/h2non/bimg#prerequisites)

## Usage
to be continued....

### predict with image crop
just use URL string with size ,like this: <br>
    `/predict/demo?size=100,100 ` <br>
if we failed  to parse the size,like you ignore this url string or having non-numberic , we will ignore it and just save it