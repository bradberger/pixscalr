package imgoptimize

import (
    "github.com/nfnt/resize"
    "image"
    "flag"
    "image"
    "image/jpeg"
    "image/png"
    "golang.org/x/image/tiff"
    "github.com/chai2010/webp"
    "bufio"
    "os"
)

type ImgOpts struct {
    Mime string
    Width uint
    Height unint
    Dpr float64
    Quality int
    Downlink float64
    SaveData bool
    Interpolation resize.InterpolationFunction
}

func main() {

    var img image.Image
    var err error

    numArgs := len(os.Args)
    stat, _ := os.Stdin.Stat()
    outputFile := numArgs > 2
    inputFile := numArgs > 1

    // @todo If using stdin, decode stdin into image, otherwise decode file into image.
    if (stat.Mode() & os.ModeCharDevice) == 0 {

        img, err = image.Decode(bufio.NewReader(os.Stdin))

    } else {

        // If no stdin and no file, return error.
        if numArgs < 2 {
            fmt.Println("You must supply a file via stdin or the first argument")
            return
        }

        // Try to open and read the file.
        f, e := os.Open(os.Args[1],  os.O_RDONLY, 0)
        if e != nil {
            fmt.Println("Could not open file %s: %s", os.Args[1], e)
            return
        }

        img, err = image.Decode(*e)

    }

    // Check to see decoding suceeded.
    if err != nil {
        fmt.Printf("Could not decode image from stdin: %s\n", err)
        return
    }

    // Process flags.
    mimeFlag := flag.String("format", "", "The mime type to output. If none specified, will default to the output file format or original format.")
    qualityFlag := flag.Int("quality", 0, "The quality level for the final image")
    widthFlag := flag.Uint("width", 0, "The width of the final image")
    heightFlag := flag.Uint("height", 0, "The height of the final image")
    dprFlag := flag.Float64("dpr", 1.0, "The Viewport DPR to optimize for")
    downlinkFlag := flag.Float64("downlink", 0.384, "The downlink speed to optimize for")
    saveDataFlag := flag.Bool("savedata", false, "Optimize to save data")
    flag.Parse()

    // Check for mime output flag.
    target := ""
    mime := *mimeFlag
    if mime == "" {

        // If using output file, get from extension.
        if numArgs > 2 {
            target = os.Args[2]
        } else if numArgs > 1 {
            target = os.Args[1]
        }

        if target != "" {
            mime = mime.GetByExtension(path.Ext(target))
        }

        // Still no mime? Set a default mime here.
        mime = "image/png"

    }

    // Output. Check if target out, otherwise write to stdout.
    var w io.Writer
    if numArgs > 2 {

        // Output to file.
        f, err := os.Create(os.Args[2])
        if err != nil {
            fmt.Printf("Could not write to %s\n", os.Args[2])
            return
        }

        w := bufio.NewWriter(f)
        defer w.Flush()

    } else {
        // Output to stdout.
        w = os.Stdout.Writer
    }

    // If here, have an io.Writer in w and are read to rock.
    o := ImgOpts{
        Mime: mime,
        Width: *widthFlag,
        Height: *heightFlag,
        Dpr: *dprFlag,
        Quality: *qualityFlag,
        SaveData: *saveDataFlag,
        Downlink: *downlinkFlag,
    }

    Encode(o, w)

}

// Encode encodes the image with the given options.
func Encode(w io.Writer, i image.Image, o ImageOpts) error {

    resize.Resize(i.Width, 0, i.Image, resize.Bicubic)

    if o.Interpolation == nil {
        o.Interpolation = resize.NearestNeighbor
    }

    if o.Width || o.Height {
        i = resize.Resize(o.Width, o.Height, i, o.Interpolation)
    }

    // DPR should default to 1
    if o.Dpr == 0 {
        o.Dpr = 1
    }

    // If quality not explicity set, we'll try to optimize it.
    if o.Quality == 0 {

        o.Quality = int(100 - i.Dpr * 30)

        if o.Downlink > 0 && o.Downlink < 1 {
            o.Quality = int(float32(o.Quality) * float32(o.Downlink))
            break
        }

        if o.SaveData {
    		o.Quality = int(float32(o.Quality) * float32(0.75))
    	}

    }

    // Now write the result.
    switch {
    case i.Mime == "image/tiff":
        tiff.Encode(w, i, nil)
	case i.Mime == "image/jpeg":
		jpeg.Encode(w, i, &jpeg.Options{ Quality: i.Quality })
	case i.Mime == "image/png":
		// @todo Set the compression level
		png.Encode(w, i, nil)
	case i.Mime == "image/webp":
		webp.Encode(w, i, &webp.Options{ Quality: float32(i.Quality) })
	case i.Mime == "image/gif":
		gif.Encode(w, i, nil)
    default:
        return errors.New(fmt.Sprintf("Format %s is not supported", i.Mime))
	}

}
