## Configuration/Command Line Options

- `--listen` A ip/hostname and port combo, like `127.0.0.1:3000` or `:3000`
- `--cache` Boolean whether to use file cache.
- `--cache-dir` The directory to cache downloaded files in. Only works if `--cache` is enabled.
- `--prefix-cdn` The prefix for serving cdn content.
- `--prefix-img` The prefix for serving external images.


## Static CDN

Static resources are served via 6 CDN's. The first response
is used:

- [cdnjs](https://cdnjs.com/)
- [JSDELIVR](http://www.jsdelivr.com/)
- [Google Hosted Libraries](https://developers.google.com/speed/libraries/)
- [OSSCDN](http://osscdn.com/#/)
- [Yandex](https://tech.yandex.ru/jslibs/)
- [Microsoft Ajax CDN](http://www.asp.net/ajax/cdn)

## Responsive Images

### Headers/Parameters ###

This section defines the headers that will be parsed and reacted on
accordingly. The server will be fully compatible with Client hints
which are available from Chrome 46 and onward.

To support browsers that don't yet have client hints, each of the available
options can also be set using a `GET` parameter:

 - `?width=100`
 - `?dpr=2.0`
 - `?viewport-width=100`
 - `?downlink=0.384`

#### `Width` ####

Per the Chrome spec, the preferred method for defining the requested
width of the image is the `Width` header.

#### `Downlink` ####

Adjusts the compression of the image based on the Downlink speed.


#### `Accept` ####

To support Chrome and WebP, the Accept header is the definitive
way to set the Content-Type of the returned image. This is the only thing
that can not be set via a `GET` parameter as it would be pointless.

#### `DPR` ####

DPR is supported. The literal size of the image returned will be `<width> * <dpr>`

Quality is also adjusted depending on the DPR, as research shows that a higher
DPR allows for more compression before noticeable degradation.

@todo Needs source


#### `Viewport-Width` ####

This is used right now to ensure that an image's width is never greater
than the viewport-width.

@todo Think there are other things to do with it. Need to check that out.


#### References ####

- [http://igrigorik.github.io/http-client-hints/](http://igrigorik.github.io/http-client-hints/)
- [https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints](https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints)
- [http://www.html5rocks.com/en/mobile/high-dpi/](http://www.html5rocks.com/en/mobile/high-dpi/)
- [http://ivomynttinen.com/blog/a-guide-for-creating-a-better-retina-web/](http://ivomynttinen.com/blog/a-guide-for-creating-a-better-retina-web/)


### Quality ###

In combination with both format and `DPR` the compression levels should be
variable to produce the best quality image which uses the lowest amount of
bandwidth possible.

#### References ####

- [http://www.netvlies.nl/blog/design-interactie/retina-revolution](http://www.netvlies.nl/blog/design-interactie/retina-revolution)
