## Design

### Scaling

Where should cached images be stored?

- In memory cache (tempfs)
- Some kind of network file system
- On local server itself

## Options ##

This section defines the options available via the user's control panel.

## Headers ##

This section defines the headers that will be parsed and reacted on
accordingly. The server will be fully compatible with Client hints
which are available from Chrome 46 and onward.

### `Width` ###

Per the Chrome spec, the preferred method for defining the requested
width of the image is the `Width` header.

### `Accept` ###

To support Chrome and WebP, the Accept header is the definitive
way to set the Content-Type of the returned image. Chrome uses:

```
Accept: image/webp,image/*,*/*
```

The server will set the `Vary` response header:

```
Vary: Accept, DPR, Width
```

#### References ####

- None, yet.

### `DPR` ###

DPR stands for "device pixel ratio" and is used to tell the server what

The server will set the `Vary` response header:

```
Vary: Accept, DPR, Width
```

#### References ####

- [http://www.html5rocks.com/en/mobile/high-dpi/](http://www.html5rocks.com/en/mobile/high-dpi/)
- [http://ivomynttinen.com/blog/a-guide-for-creating-a-better-retina-web/](http://ivomynttinen.com/blog/a-guide-for-creating-a-better-retina-web/)


### `Viewport-Width` ###

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

## Caching

This section defines the caching process.

## Formats

This section defines the various formats that will be available.

- WebP (preferred)
- JPEG
- PNG
- GIF
- TIFF (read-only, never served)


## Compression

In combination with both format and `DPR` the compression levels should be
variable to produce the best quality image which uses the lowest amount of
bandwidth possible.

#### References

- [http://www.netvlies.nl/blog/design-interactie/retina-revolution](http://www.netvlies.nl/blog/design-interactie/retina-revolution)
