package udf

import (
	"fmt"
	"io"
	"os"
	"time"
)

type File struct {
	Udf               *Udf
	Fid               *FileIdentifierDescriptor
	fe                *FileEntry
	fileEntryPosition uint64
}

func (f *File) GetFileEntryPosition() int64 {
	return int64(f.fileEntryPosition)
}

func (f *File) GetFileOffset() (int64, error) {
	if len(f.FileEntry().AllocationDescriptors) == 0 {
		return 0, fmt.Errorf("empty descriptors")
	}
	return SECTOR_SIZE * (int64(f.FileEntry().AllocationDescriptors[0].Location) + int64(f.Udf.PartitionStart())), nil
}

func (f *File) FileEntry() *FileEntry {
	if f.fe == nil {
		f.fileEntryPosition = f.Fid.ICB.Location
		f.fe = NewFileEntry(f.Udf.ReadSector(f.Udf.PartitionStart() + f.fileEntryPosition))
	}
	return f.fe
}

func (f *File) NewReader() (sr *io.SectionReader, err error) {

	offset, err := f.GetFileOffset()
	if err != nil {
		return
	}
	return io.NewSectionReader(f.Udf.r, offset, f.Size()), nil

}

func (f *File) Name() string {
	return f.Fid.FileIdentifier
}

func (f *File) Mode() os.FileMode {
	var mode os.FileMode

	perms := os.FileMode(f.FileEntry().Permissions)
	mode |= ((perms >> 0) & 7) << 0
	mode |= ((perms >> 5) & 7) << 3
	mode |= ((perms >> 10) & 7) << 6

	if f.IsDir() {
		mode |= os.ModeDir
	}

	return mode
}

func (f *File) Size() int64 {
	return int64(f.FileEntry().InformationLength)
}

func (f *File) ModTime() time.Time {
	return f.FileEntry().ModificationTime
}

func (f *File) IsDir() bool {
	// TODO :Fix! This field always 0 :(
	return f.FileEntry().ICBTag.FileType == 4
}

func (f *File) Sys() interface{} {
	return f.Fid
}

func (f *File) ReadDir() ([]File, error) {
	return f.Udf.ReadDir(f.FileEntry())
}
