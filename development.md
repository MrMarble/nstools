# [PFS0](https://switchbrew.org/wiki/NCA#PFS0)

Container format for storing multiple files.

```cpp
import std.string;

char magic[0x04] @ 0x00;
u32 file_count @ 0x04;
u32 string_table_size @ 0x08;

u32 string_table_offset = 0x10 + (file_count * 0x18);
u32 header_size = string_table_offset + string_table_size;

struct File {
  u64 file_offset;
  u64 file_size;
  u32 name_offset;
  padding[4];
  std::string::NullString name @ name_offset + string_table_offset;
};

File files[file_count] @ 0x10;
```
