#pragma MIME application/ponzu-archive
#pragma MIME application/x-ponzu-archive

#include <std/string.pat>
#include <std/time.pat>
#include <std/mem.pat>
#include <std/sys.pat>


#define RECORD_MAGIC "PONZU\x00"
#define BLOCK_SIZE 4096

#pragma endian big

enum recordType :u8 {
    Control,
    File,
    Symlink,
    Hardlink,
    Directory,
    ZstdDict,
    OSSpecial = 126,
    ContinueBlock = 127
};

enum CompressionType : u8 {
    None,
    ZStandard,
    Brotli
};

enum Flags : u16 {
    none = 0,
    StartArchive = 1,
    EndArchive =2,
    ArchiveContinues = 1
};

struct Preamble {
    char magic[6];
    std::print("Magic: {}", magic);
    std::assert(magic == RECORD_MAGIC, "Bad preamble magic. Wrong file?");
    recordType type;
    CompressionType compression;
    Flags flags;
    u64 blockCount [[name("Full blocks")]];
    u16 dataModulo [[name("Last block usage")]];
    u8 checksum[64] [[name("Body BLAKE2b")]];
    u16 metadataSize [[name("Metadata Size")]];
    u8 metadataChecksum[64] [[name("Metadata Blake2b")]];

};


struct Record {

    Preamble preamble;
    u8 metadataContent[preamble.metadataSize] [[name("metadata CBOR")]];
    std::print("padding={}", BLOCK_SIZE - ( sizeof(preamble) + preamble.metadataSize) );
    padding[BLOCK_SIZE -  ( sizeof(preamble) + preamble.metadataSize) ];
    std::print("type {}, size {}", preamble.type, preamble.blockCount);
    std::print("Metadata size: {}",preamble.metadataSize);
    if (preamble.blockCount > 0) {
        u8 body[((preamble.blockCount-1)*BLOCK_SIZE)+preamble.dataModulo];
        padding[BLOCK_SIZE-preamble.dataModulo];
    }
};


Record archive[while(!std::mem::eof())] @ 0x00 [[name("Ponzu Archive")]];