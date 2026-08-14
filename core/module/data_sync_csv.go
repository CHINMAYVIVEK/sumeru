package module

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/orm"
)

func (addon *Addon) syncCSVModelAccess(ctx context.Context) error {
	csvPath := filepath.Join(addon.Path, "sys.access.csv")
	if _, err := os.Stat(csvPath); err != nil {
		csvPath = filepath.Join(addon.Path, "security", "sys.access.csv")
		if _, err := os.Stat(csvPath); err != nil {
			return nil // No CSV ACL file found
		}
	}

	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	// Skip header: id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
	if _, err := csvReader.Read(); err != nil {
		return err
	}

	for {
		csvRecord, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(csvRecord) < 8 {
			continue
		}

		recordXmlId := csvRecord[0]
		accessName := csvRecord[1]
		modelName := csvRecord[2]
		groupXmlId := csvRecord[3]
		permRead := csvRecord[4] == "1"
		permWrite := csvRecord[5] == "1"
		permCreate := csvRecord[6] == "1"
		permUnlink := csvRecord[7] == "1"

		var groupId int
		if groupXmlId != "" {
			gid, _, err := orm.ResolveXmlId(ctx, groupXmlId)
			if err != nil {
				// Try with module prefix if not absolute
				if !strings.Contains(groupXmlId, ".") {
					gid, _, _ = orm.ResolveXmlId(ctx, addon.Manifest.Name+"."+groupXmlId)
				}
			}
			groupId = gid
		}

		accessValues := map[string]interface{}{
			"name":        accessName,
			"model":       modelName,
			"perm_read":   permRead,
			"perm_write":  permWrite,
			"perm_create": permCreate,
			"perm_unlink": permUnlink,
		}
		if groupId > 0 {
			accessValues["group_id"] = groupId
		}

		id, err := orm.Upsert(ctx, orm.SysAccess{}, accessValues, "name")
		if err == nil {
			_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
				"module":  addon.Manifest.Name,
				"name":    recordXmlId,
				"model":   "sys.access",
				"core_id": id,
			}, "name")
		}
	}
	return nil
}
