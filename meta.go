package meta

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-meta/meta"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type WOFMetaDefaults map[string]interface{}

func (d WOFMetaDefaults) EnsureDefaults(row map[string]string) map[string]string {

	out := make(map[string]string)

	for k, v := range row {

		if v != "" {
			out[k] = v
			continue
		}

		default_v, ok := d[k]

		if !ok {
			out[k] = v
			continue
		}

		out[k] = default_v.(string)
	}

	for k, default_v := range d {

		_, ok := out[k]

		if !ok {
			out[k] = default_v.(string)
		}
	}

	return out
}

var defaults *WOFMetaDefaults

func init() {

	var err error

	defaults, err = GetDefaults()

	if err != nil {
		log.Fatal("Failed to parse defaults", err)
	}
}

func GetDefaults() (*WOFMetaDefaults, error) {

	var defaults WOFMetaDefaults
	err := json.Unmarshal([]byte(meta.Defaults), &defaults)

	if err != nil {
		return nil, err
	}

	return &defaults, nil
}

func UpdateMetafile(source io.Reader, dest io.Writer, updated []string) error {

	lookup := make(map[int64][]byte)

	for _, path := range updated {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			return err
		}

		fh, err := os.Open(abs_path)

		if err != nil {
			return err
		}

		feature, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
		}

		wofid_fl := gjson.GetBytes(feature, "properties.wof:id").Float()
		wofid := int64(wofid_fl)

		lookup[wofid] = feature
	}

	var writer *csv.DictWriter

	reader, reader_err := csv.NewDictReader(source)

	if reader_err != nil {
		return reader_err
	}

	for {
		row, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		str_wofid, ok := row["id"]
		wofid, err := strconv.Atoi(str_wofid)

		if err != nil {
			return err
		}

		feature, ok := lookup[int64(wofid)]

		if ok {

			new_row, err := DumpFeature(feature)

			if err != nil {
				return err
			}

			row = defaults.EnsureDefaults(new_row)
		}

		if writer == nil {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			writer, err = csv.NewDictWriter(dest, fieldnames)

			if err != nil {
				return err
			}

			writer.WriteHeader()
		}

		writer.WriteRow(row)
	}

	return nil
}

func DumpFeature(feature []byte) (map[string]string, error) {

	row := make(map[string]string)

	wofid_fl := gjson.GetBytes(feature, "properties.wof:id").Float()
	wofid := int64(wofid_fl)

	rel_path, err := uri.Id2RelPath(int(wofid))

	if err != nil {
		return row, err
	}

	row["id"] = strconv.FormatInt(wofid, 10)
	row["path"] = rel_path

	row["name"] = gjson.GetBytes(feature, "properties.wof:name").String()
	// row["properties"] = gjson.GetBytes(feature, "properties.wof:properties").String()

	row["source"] = gjson.GetBytes(feature, "properties.src:geom").String()

	// TO DO
	// bbox geom:bbox

	supersedes := make([]string, 0)
	superseded_by := make([]string, 0)

	for _, r := range gjson.GetBytes(feature, "properties.wof:supersedes").Array() {

		wofid_fl := r.Float()
		wofid := int64(wofid_fl)
		id_str := strconv.FormatInt(wofid, 10)

		supersedes = append(supersedes, id_str)
	}

	row["supersedes"] = strings.Join(supersedes, ",")

	for _, r := range gjson.GetBytes(feature, "properties.wof:superseded_by").Array() {

		wofid_fl := r.Float()
		wofid := int64(wofid_fl)
		id_str := strconv.FormatInt(wofid, 10)

		supersedes = append(supersedes, id_str)
	}

	row["superseded_by"] = strings.Join(superseded_by, ",")

	row["iso"] = gjson.GetBytes(feature, "properties.iso:country").String()
	row["iso_country"] = row["iso"]
	row["wof_country"] = gjson.GetBytes(feature, "properties.wof:country").String()

	lastmod_fl := gjson.GetBytes(feature, "properties.wof:lastmodified").Float()
	lastmod := int(lastmod_fl)

	row["lastmodifed"] = strconv.Itoa(lastmod)

	row["geom_hash"] = gjson.GetBytes(feature, "properties.geom:hash").String()

	geom_lat := gjson.GetBytes(feature, "properties.geom:latitude").Float()
	geom_lon := gjson.GetBytes(feature, "properties.geom:longitude").Float()

	str_geom_lat := strconv.FormatFloat(geom_lat, 'f', 6, 64)
	str_geom_lon := strconv.FormatFloat(geom_lon, 'f', 6, 64)

	row["geom_latitude"] = str_geom_lat
	row["geom_longitude"] = str_geom_lon

	lbl_lat := gjson.GetBytes(feature, "properties.lbl:latitude").Float()
	lbl_lon := gjson.GetBytes(feature, "properties.lbl:longitude").Float()

	str_lbl_lat := strconv.FormatFloat(lbl_lat, 'f', 6, 64)
	str_lbl_lon := strconv.FormatFloat(lbl_lon, 'f', 6, 64)

	row["lbl_latitude"] = str_lbl_lat
	row["lbl_longitude"] = str_lbl_lon

	row["inception"] = gjson.GetBytes(feature, "properties.edtf:inception").String()
	row["cessation"] = gjson.GetBytes(feature, "properties.edtf:cessation").String()
	row["deprecated"] = gjson.GetBytes(feature, "properties.edtf:deprecated").String()

	parent_fl := gjson.GetBytes(feature, "properties.wof:parent_id").Float()
	parent_id := int64(parent_fl)

	row["parent_id"] = strconv.FormatInt(parent_id, 10)

	country_fl := gjson.GetBytes(feature, "properties.wof:hierarchy.0.country_id").Float()
	country_id := int64(country_fl)

	region_fl := gjson.GetBytes(feature, "properties.wof:hierarchy.0.region_id").Float()
	region_id := int64(region_fl)

	locality_fl := gjson.GetBytes(feature, "properties.wof:hierarchy.0.locality_id").Float()
	locality_id := int64(locality_fl)

	row["country_id"] = strconv.FormatInt(country_id, 10)
	row["region_id"] = strconv.FormatInt(region_id, 10)
	row["locality_id"] = strconv.FormatInt(locality_id, 10)

	return row, nil
}
