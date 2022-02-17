package slurmjob

import (
        "io/ioutil"
        "os"
        "testing"
)

func TestParseSstatMetrics(t *testing.T) {

}

func TestParseSacctMetrics(t *testing.T) {
        // Read the input data from a file
        file, err := os.Open("../../test_data/sstat.txt")
        if err != nil {
                t.Fatalf("Can not open test data: %v", err)
        }
        data, _ := ioutil.ReadAll(file)
        metrics := ParseSstatMetrics(data)
        t.Logf("%+v", metrics)
        if metrics.MaxRSS != 1850245120 {
                t.Errorf("MaxRSS is incorrect. got: %d, want: %d", metrics.MaxRSS, 1850245120)
        }
        if metrics.MaxDiskWrite != 70 {
                t.Errorf("MaxDiskWrite is incorrect. got: %d, want: %d", metrics.MaxDiskWrite, 70)
        }
        if metrics.MaxDiskRead != 205384 {
                t.Errorf("MaxDiskRead is incorrect. got: %d, want: %d", metrics.MaxDiskRead, 205384)
        }

}
