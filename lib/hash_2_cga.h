#include <stddef.h>

typedef unsigned char u8;
typedef unsigned int  u32;

const u8* generate_hash_2(const u32 threads, const u8 sec, const size_t pub_key_size, const u8* pub_key);
