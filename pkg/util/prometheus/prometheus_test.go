package prometheus

import (
	"bytes"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metrics "kubevirt.io/containerized-data-importer/pkg/monitoring/metrics/cdi-cloner"
	"kubevirt.io/containerized-data-importer/pkg/util"
)

const (
	ownerUID = "1111-1111-111"
)

var _ = Describe("Timed update", func() {

	It("Should start and stop when finished", func() {
		r := io.NopCloser(bytes.NewReader([]byte("hello world")))
		progressReader := NewProgressReader(r, uint64(11), ownerUID)
		progressReader.StartTimedUpdate()
		_, err := io.ReadAll(r)
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("Update Progress", func() {
	BeforeEach(func() {
		err := metrics.SetupMetrics()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		metrics.DeleteCloneProgress(ownerUID)
	})

	It("Parse valid progress update", func() {
		By("Verifying the initial value is 0")
		metrics.AddCloneProgress(ownerUID, 0)
		progress, err := metrics.GetCloneProgress(ownerUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(progress).To(Equal(float64(0)))
		By("Calling updateProgress with value")
		promReader := &ProgressReader{
			CountingReader: util.CountingReader{
				Current: uint64(45),
			},
			total:    uint64(100),
			ownerUID: ownerUID,
			final:    true,
		}
		result := promReader.updateProgress()
		Expect(true).To(Equal(result))
		progress, err = metrics.GetCloneProgress(ownerUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(progress).To(Equal(float64(45)))
	})

	It("0 total should return 0", func() {
		By("Calling updateProgress with value")
		promReader := &ProgressReader{
			CountingReader: util.CountingReader{
				Current: uint64(45),
			},
			total:    uint64(0),
			ownerUID: ownerUID,
			final:    true,
		}
		result := promReader.updateProgress()
		Expect(false).To(Equal(result))
		progress, err := metrics.GetCloneProgress(ownerUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(progress).To(Equal(float64(0)))
	})

	It("current and total equals should return false", func() {
		By("Calling updateProgress with value")
		promReader := &ProgressReader{
			CountingReader: util.CountingReader{
				Current: uint64(1000),
				Done:    true,
			},
			total:    uint64(1000),
			ownerUID: ownerUID,
			final:    true,
		}
		result := promReader.updateProgress()
		Expect(false).To(Equal(result))
		progress, err := metrics.GetCloneProgress(ownerUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(progress).To(Equal(float64(100)))
	})

	DescribeTable("update progress on non-final readers", func(readerDone, isFinal, expectedResult bool) {
		promReader := &ProgressReader{
			CountingReader: util.CountingReader{
				Current: uint64(1000),
				Done:    readerDone,
			},
			total:    uint64(1000),
			ownerUID: ownerUID,
			final:    isFinal,
		}
		result := promReader.updateProgress()
		Expect(expectedResult).To(Equal(result))
	},
		Entry("should return true when reader is not done", false, false, true),
		Entry("should return true when reader is done", true, false, true),
		Entry("should return true when final reader is not done", false, true, true),
		Entry("should return false when final reader is done", true, true, false),
	)

	It("should continue to update progress after next reader is set", func() {
		firstReader := util.CountingReader{
			Reader: io.NopCloser(strings.NewReader("first")),
		}
		secondReader := util.CountingReader{
			Reader: io.NopCloser(strings.NewReader("second")),
		}
		thirdReader := util.CountingReader{
			Reader: io.NopCloser(strings.NewReader("third")),
		}
		promReader := &ProgressReader{
			CountingReader: firstReader,
			total:          uint64(16),
			ownerUID:       ownerUID,
			final:          false,
		}

		data := make([]byte, 10)
		read, _ := promReader.Read(data)
		Expect(read).To(Equal(5))
		_, err := promReader.Read(data)
		Expect(err).To(Equal(io.EOF))
		result := promReader.updateProgress()
		Expect(true).To(Equal(result))
		Expect(promReader.CountingReader.Current).To(Equal(uint64(5)))

		promReader.SetNextReader(secondReader.Reader, false)
		read, _ = promReader.Read(data)
		Expect(read).To(Equal(6))
		_, err = promReader.Read(data)
		Expect(err).To(Equal(io.EOF))
		result = promReader.updateProgress()
		Expect(promReader.CountingReader.Reader).To(Equal(secondReader.Reader))
		Expect(promReader.CountingReader.Current).To(Equal(uint64(11)))
		Expect(true).To(Equal(result))

		promReader.SetNextReader(thirdReader.Reader, true)
		read, _ = promReader.Read(data)
		Expect(read).To(Equal(5))
		_, err = promReader.Read(data)
		Expect(err).To(Equal(io.EOF))
		result = promReader.updateProgress()
		Expect(promReader.CountingReader.Reader).To(Equal(thirdReader.Reader))
		Expect(promReader.CountingReader.Current).To(Equal(uint64(16)))
		Expect(false).To(Equal(result))
	})
})
