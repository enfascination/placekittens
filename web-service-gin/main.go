package main

import (
	"fmt"
	"image"
	"image/color"
	//"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"github.com/gin-gonic/gin"
	//"github.com/nfnt/resize"
	"github.com/disintegration/imaging"
)

// GetRandomFile returns the filename of a random file in the specified directory
func GetRandomImage(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, filepath.Join(dir, file.Name()))
		}
	}

	if len(fileList) == 0 {
		return "", fmt.Errorf("no files found in directory")
	}

	randomIndex := rand.Intn(len(fileList))

	return fileList[randomIndex], nil
}

// GetApproxImage returns the filename of the image in dir whose aspect ratio is
// closest to width:height. If multiple images are equally close, one is chosen randomly.
func GetApproxImage(dir string, width, height int) (string, error) {
        targetRatio := float64(width) / float64(height)

        files, err := os.ReadDir(dir)
        if err != nil {
                return "", err
        }

        type fileAndDiff struct {
                filename string
                diff     float64
        }

        var candidates []fileAndDiff
        minDiff := math.MaxFloat64
	
        for _, file := range files {
                if !file.IsDir() {
                        fullPath := filepath.Join(dir, file.Name())
                        f, err := os.Open(fullPath)
                        if err != nil {
                                continue
                        }
                        config, _, err := image.DecodeConfig(f)
                        f.Close()
                        if err != nil {
                                continue // Skip non-image files, or unreadable
                        }
                        aspect := float64(config.Width) / float64(config.Height)
                        diff := math.Abs(aspect - targetRatio)
                        if diff < minDiff {
                                minDiff = diff
                                candidates = candidates[:0] // Reset candidates
                                candidates = append(candidates, fileAndDiff{filename: fullPath, diff: diff})
                        } else if diff == minDiff {
                                candidates = append(candidates, fileAndDiff{filename: fullPath, diff: diff})
                        }
                }
        }

        if len(candidates) == 0 {
                return "", fmt.Errorf("no image files found in directory")
        }
        // Randomly pick among the best matches.
        idx := rand.Intn(len(candidates))
        return candidates[idx].filename, nil
}	

// This function should return a resized color image file
func returnColorImage(c *gin.Context) {
	w := c.Param("width")
	h := c.Param("height")

	height, err := strconv.Atoi(h)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid height"})
		return
	}

	width, err := strconv.Atoi(w)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid width"})
		return
	}

    //filename, err := GetRandomImage("../public")
    filename, err := GetApproxImage("../public", width, height)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to find image"})
		return
	}

	filepath := "../public/" + filename

	file, err := os.Open(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open image"})
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to decode image"})
		return
	}

	//resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
    resizedImg := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

    //c.Writer.Header().Set("Content-Type", "image/jpeg")
    //jpeg.Encode(c.Writer, resizedImg, nil)
    c.Writer.Header().Set("Content-Type", "image/png")
    png.Encode(c.Writer, resizedImg)
}

func returnGreyImage(c *gin.Context) {
	w := c.Param("width")
	h := c.Param("height")

	height, err := strconv.Atoi(h)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid height"})
		return
	}

	width, err := strconv.Atoi(w)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid width"})
		return
	}

	//filename, err := GetRandomImage("../public")
	filename, err := GetApproxImage("../public", width, height)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to find image"})
		return
	}

	filepath := "../public/" + filename

	file, err := os.Open(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open image"})
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to decode image"})
		return
	}

	//resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
	resizedImg := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	grayImg := image.NewGray(resizedImg.Bounds())
	for y := 0; y < resizedImg.Bounds().Dy(); y++ {
		for x := 0; x < resizedImg.Bounds().Dx(); x++ {
			originalColor := resizedImg.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor)
			grayImg.Set(x, y, grayColor)
		}
	}

	//c.Writer.Header().Set("Content-Type", "image/jpeg")
	//jpeg.Encode(c.Writer, grayImg, nil)
	c.Writer.Header().Set("Content-Type", "image/png")
    png.Encode(c.Writer, resizedImg)
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("./static/*.html")
    router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })

	router.GET("/:width/:height", returnColorImage)
	router.GET("/g/:width/:height", returnGreyImage)
	router.Run("localhost:8080")
}
