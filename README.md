# lyon
Lyon+Post Challenge

### Quick Usage
1. Ensure golang is installed
2. From the challenge directory execute **go run main.go**

### Advanced Usage
Execute **go run main.go -h** to display full usage.<br>
Replace **challenge** with **go run main.go** to avoid installing with **go install**.

### Implementation Notes
The code is split into 3 parts. Importer loads the jpegs from the api/local directory and either 
directly sends the data along for processing or saves it to a local directory. Exporter writes results 
to file using a buffer. Processor uses the color classification file to create color 'buckets'. It converts 
all colors (buckets and data) to CIELab using go-chromath.<br>

The processor takes a central subset of the image defined by the constant BOUNDS_FRACTION. Depends on the
assumption that important pixels are in the central region. For each pixel in the region, it converts the
color to Lab, matches it to a bucket using chromath and increments a counter for that bucket. Once all pixels
are processed, return the buckets with the top counts. Only buckets with counts above a threshold defined by
the constant PX_THRESHOLD_RATIO are returned. If no buckets break the threshold, image is labeled multicolored.
White pixels have a stricter threshold defined by PX_THRESHOLD_RATIO_WHITE.

### TODO
1. Everything's set up to have processor spawn workers to process images in parallel across CPU cores. Ran out
of time but some adjustments to the semaphore in SyncData and processor's process() method (few other changes)
would do it.
2. Would have been nice to sample a subset of pixels rather than process all of them within the inner bounds.
Have to be satisfied with only processing a smaller window within the image for now.
