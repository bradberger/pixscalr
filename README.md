<<<<<<< HEAD
## Static CDN

Static resources are served via 6 CDN's. The first response
is used:

- [cdnjs](https://cdnjs.com/)
- [JSDELIVR](http://www.jsdelivr.com/)
- [Google Hosted Libraries](https://developers.google.com/speed/libraries/)
- [OSSCDN](http://osscdn.com/#/)
- [Yandex](https://tech.yandex.ru/jslibs/)
- [Microsoft Ajax CDN](http://www.asp.net/ajax/cdn)

The


## Responsive Images

### Headers/Parameters ###
=======
## Headers/Parameters ##
>>>>>>> 906b1e6c7012bd8558d5bd32f4df827cf01cff2a

This section defines the headers that will be parsed and reacted on
accordingly. The server will be fully compatible with Client hints
which are available from Chrome 46 and onward.

To support browsers that don't yet have client hints, each of the available
options can also be set using a `GET` parameter:

 - `?width=100`
 - `?dpr=2.0`
 - `?viewport-width=100`
 - `?downlink=0.384`

<<<<<<< HEAD
#### `Width` ####
=======
### `Width` ###
>>>>>>> 906b1e6c7012bd8558d5bd32f4df827cf01cff2a

Per the Chrome spec, the preferred method for defining the requested
width of the image is the `Width` header.

<<<<<<< HEAD
#### `Downlink` ####

Adjusts the compression of the image based on the Downlink speed.

#### `Accept` ####
=======
### `Downlink` ###

Adjusts the compression of the image based on the Downlink speed.

### `Accept` ###
>>>>>>> 906b1e6c7012bd8558d5bd32f4df827cf01cff2a

To support Chrome and WebP, the Accept header is the definitive
way to set the Content-Type of the returned image. This is the only thing
that can not be set via a `GET` parameter as it would be pointless.

<<<<<<< HEAD
#### `DPR` ####
=======
### `DPR` ###
>>>>>>> 906b1e6c7012bd8558d5bd32f4df827cf01cff2a

DPR is supported. The literal size of the image returned will be `<width> * <dpr>`

Quality is also adjusted depending on the DPR, as research shows that a higher
DPR allows for more compression before noticeable degradation.

@todo Needs source

<<<<<<< HEAD
#### `Viewport-Width` ####

This is used right now to ensure that an image's width is never greater
than the viewport-width.

@todo Think there are other things to do with it. Need to check that out.

### Response Headers ###
=======
### `Viewport-Width` ###

This is used right now to ensure that an image's width is never greater
than the viewport-width.

@todo Think there are other things to do with it. Need to check that out.

## Response Headers
>>>>>>> 906b1e6c7012bd8558d5bd32f4df827cf01cff2a

The server will set the `Vary` response header:

```
Vary: Accept, DPR, Width, Save-Data, Downlink
```

#### References ####

- [http://igrigorik.github.io/http-client-hints/](http://igrigorik.github.io/http-client-hints/)
- [https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints](https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints)
- [http://www.html5rocks.com/en/mobile/high-dpi/](http://www.html5rocks.com/en/mobile/high-dpi/)
- [http://ivomynttinen.com/blog/a-guide-for-creating-a-better-retina-web/](http://ivomynttinen.com/blog/a-guide-for-creating-a-better-retina-web/)




The server will set the `Vary` response header:

```
Vary: Accept, DPR, Width
```

#### References ####

- None, yet.

- [Automating resource selection with Client Hints](https://developers.google.com/web/updates/2015/09/automating-resource-selection-with-client-hints)

## Request parameters

This section defines the request parameters available for client side
requests to use/send.  

### Width ###

Same implementation as the `Viewport-Width` header, just for clients that
don't send the header.

#### Ways to Set It ####

- `width` The `width` query parameter will set width if header information is
  not available.

- `/<w>/path/to/image` If the first value here is numeric, it will become the
  request width for the image. This is to allow better caching upstream by
  caching proxies, and services such as CloudFlare, which may ignore query
  parameters.

### Viewport-Width ###

Same implementation as the `Viewport-Width` header, just for clients that
don't send the header.

### DPR ###

Same implementation as the `DPR` header, just for clients that don't send
the header.

### Caching ###

This section defines the caching process.

### Formats ###

This section defines the various formats that will be available.

- WebP (preferred)
- JPEG
- PNG
- GIF
- TIFF (read-only, never served)


### Compression ###

In combination with both format and `DPR` the compression levels should be
variable to produce the best quality image which uses the lowest amount of
bandwidth possible.

#### References ####

- [http://www.netvlies.nl/blog/design-interactie/retina-revolution](http://www.netvlies.nl/blog/design-interactie/retina-revolution)
