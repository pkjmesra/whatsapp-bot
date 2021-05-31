package pkBot

import (
  "encoding/json"
  "fmt"
  "image"
  "image/png"
  "image/color"
  "image/draw"
  "image/jpeg"
  "log"
  "os"
  
  // "io"
  // "html"

  // "bytes"
  // // "flag"
  // // "fmt"
  // // "net/http"
  // // "net/http/cookiejar"
  // // "net/url"
  // // "os"
  // // "strings"

  // tf "github.com/galeone/tensorflow/tensorflow/go"
  
  // "github.com/otiai10/gosseract/v2"
  
  "github.com/srwiley/oksvg"
  "github.com/srwiley/rasterx"
)

import _ "image/png"    //register PNG decoder

type CaptchaSVG struct {
  Captcha string `json:"captcha"`
}

func getCaptchaSVG(svgFileName string) error {
  log.Print("Generating SVG CAPTCHA")
  postBody := map[string]interface{}{}
  response, err := queryServer(captchaURLFormat, "POST", postBody)
  
  if err != nil {
     fmt.Println(err)
     return err
  }
  ctcha := CaptchaSVG{}
  if err := json.Unmarshal(response, &ctcha); err != nil {
    log.Printf("Error while parsing:%s",err.Error())
    return err
  }
  svgFile, err := os.Create(pkBotFilePath(svgFileName))
  if err != nil {
     fmt.Println(err)
     return err
  }
  defer svgFile.Close()
  svgBytes := []byte(ctcha.Captcha)
  _ , err = svgFile.Write(svgBytes)
  if err != nil {
     fmt.Println(err)
     return err
  }
  fmt.Println("SVG data saved into :" + pkBotFilePath(svgFileName))
  return nil
}

func exportToPng(svgFileName string, pngFileName string) error {
  w, h := 512, 512
  in, err := os.Open(pkBotFilePath(svgFileName))
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer in.Close()

  icon, _ := oksvg.ReadIconStream(in)
  w = int(icon.ViewBox.W) 
  h = int(icon.ViewBox.H)
  icon.SetTarget(0, 0, float64(w), float64(h))
  rgba := image.NewRGBA(image.Rect(0, 0, w, h))
  icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 2)

  out, err := os.Create(pkBotFilePath(pngFileName))
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer out.Close()
  err = png.Encode(out, rgba)
  if err != nil {
    fmt.Println(err)
    return err
  }
  fmt.Println("PNG data saved into :" + pkBotFilePath(pngFileName))
  return nil
}

func getPngImage(pngFileName string) (image.Image, string, string, error) {
  f, err := os.Open(pkBotFilePath(pngFileName))
  if err != nil {
    log.Printf("Error while loading PNG image:%s",err.Error())
  }
  defer f.Close()

  img, fmtName, err := image.Decode(f)
  if err != nil {
    log.Printf("Error while decoding PNG image:%s",err.Error())
  }
  
  return img, pkBotFilePath(pngFileName), fmtName, err
}

func getJPEGImage(pngFileName string) (string, error) {
   pngImgFile, err := os.Open(pkBotFilePath(pngFileName))
   if err != nil {
     fmt.Println("File not found:" + pngFileName)
     return "", err
   }
   defer pngImgFile.Close()
   // create image from PNG file
   imgSrc, err := png.Decode(pngImgFile)
   if err != nil {
     fmt.Println(err)
     return "", err
   }
   // create a new Image with the same dimension of PNG image
   newImg := image.NewRGBA(imgSrc.Bounds())
   // we will use white background to replace PNG's transparent background
   // you can change it to whichever color you want with
   // a new color.RGBA{} and use image.NewUniform(color.RGBA{<fill in color>}) function
   draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
   // paste PNG image OVER to newImage
   draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
   // create new out JPEG file
   jpgImgFile, err := os.Create(pkBotFilePath(pngFileName + ".jpg"))
   if err != nil {
     fmt.Println("Cannot create JPEG-file!")
     fmt.Println(err)
     return "", err
   }
   defer jpgImgFile.Close()
   var opt jpeg.Options
   opt.Quality = 80
   // convert newImage to JPEG encoded byte and save to jpgImgFile
   // with quality = 80
   err = jpeg.Encode(jpgImgFile, newImg, &opt)
   //err = jpeg.Encode(jpgImgFile, newImg, nil) -- use nil if ignore quality options
   if err != nil {
       fmt.Println(err)
       return "", err
   }
   fmt.Println("Converted PNG file to JPEG file")
   return pkBotFilePath(pngFileName + ".jpg"), nil
 }

// func convertToGrayScale(img image.Image) (image.Image, string, error) {
// b := img.Bounds()
// imgSet := image.NewRGBA(b)
// for y := 0; y < b.Max.Y; y++ {
//     for x := 0; x < b.Max.X; x++ {
//       imgSet.Set(x, y, color.Gray16Model.Convert(img.At(x, y)))
//      }
//     }
//     outFile, err := os.Create("captcha_grayscale.png")
//     if err != nil {
//       log.Fatal(err)
//     }
//     defer outFile.Close()
//     png.Encode(outFile, imgSet)
//     return getPngImage("captcha_grayscale.png")
// }

// func breakCaptchaTensorflow(img image.Image) string {
//   // img_bytes := buf.Bytes()
// // load tensorflow model
//   savedModel, err := tf.LoadSavedModel("./tensorflow_savedmodel_captcha", []string{"serve"}, nil)
//   if err != nil {
//     log.Println("failed to load model", err)
//     return ""
//   }
//   buf := new(bytes.Buffer)
//   err = png.Encode(buf, img)

//   // run captcha through tensorflow model
//   feedsOutput := tf.Output{
//     Op:    savedModel.Graph.Operation("CAPTCHA/input_image_as_bytes"),
//     Index: 0,
//   }
//   feedsTensor, err := tf.NewTensor(string(buf.String()))
//   if err != nil {
//     log.Fatal(err)
//     return ""
//   }
//   feeds := map[tf.Output]*tf.Tensor{feedsOutput: feedsTensor}

//   fetches := []tf.Output{
//     {
//       Op:    savedModel.Graph.Operation("CAPTCHA/prediction"),
//       Index: 0,
//     },
//   }

//   captchaText, err := savedModel.Session.Run(feeds, fetches, nil)
//   if err != nil {
//     log.Fatal(err)
//     return ""
//   }
//   captchaString := captchaText[0].Value().(string)
//   return captchaString
// }

// func resolveCaptchaTesseract(fileName string) {
//   client := gosseract.NewClient()
//   defer client.Close()
//   client.SetImage(fileName)
//   text, _ := client.Text()
//   fmt.Println("CAPTCHA from tesseract:",text)
// }