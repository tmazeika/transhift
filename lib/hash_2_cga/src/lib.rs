#![crate_type = "dylib"]

use std::sync::{Arc, RwLock};
use std::thread;
use rand::{thread_rng, Rng};

extern crate libc;
extern crate sha1;
extern crate rand;

// #[test]
// fn it_works() {
//     let hash_2 = generate_hash_2(4, 1, &[
//         0x9A, 0x58, 0x3D, 0xC1, 0xDF, 0x3D, 0xDF, 0x97, 0x33, 0xFC,
//         0xCF, 0xEC, 0x99, 0x66, 0xFF, 0x9D, 0xC3, 0xA0, 0xAF, 0x44,
//         0x16, 0xA9, 0x34, 0x56, 0x90, 0xC1, 0x00, 0xB7, 0x94, 0x7E,
//         0x64, 0x76, 0xCF, 0x1F, 0x42, 0x66, 0x80, 0x57, 0x33, 0xD5,
//         0x9E, 0x5A, 0xA3, 0xF3, 0x79, 0x4D, 0x1B, 0x7B, 0x1F, 0x58,
//         0xCE, 0x1D, 0xA4, 0x06, 0x89, 0xAB, 0x1C, 0xB6, 0x3F, 0x24,
//         0x8C, 0x8F, 0x34, 0x0D, 0xEF, 0xF7, 0x9F, 0xD2, 0xAF, 0xFF,
//         0x09, 0xBF, 0xD1, 0x31, 0x8C, 0x02, 0x62, 0x3F, 0xDF, 0xFD,
//         0xD4, 0xE0, 0xFD, 0x4F, 0xD2, 0x8A, 0xBF, 0xEE, 0x5A, 0xFC,
//         0x51, 0xEF, 0xCC, 0x9B, 0x8F, 0x24, 0xF1, 0x01, 0xEA, 0x9C,
//         0x08, 0x61, 0xA1, 0x79, 0x57, 0x2D, 0xDD, 0xD7, 0x31, 0xDB,
//         0xF1, 0x31, 0x3F, 0x39, 0x5B, 0x21, 0x1E, 0xA6, 0xAF, 0x52,
//         0x60, 0xE6, 0x6D, 0xF3, 0x0A, 0x41, 0xA4, 0x12, 0x3E, 0x73,
//         0x26, 0xB0, 0xE2, 0xEA, 0x27, 0x11, 0x97, 0x68, 0xF7, 0x1B,
//         0x67, 0x08, 0xD8, 0x2F, 0x93, 0x36, 0x2E, 0x2E, 0x67, 0x4A,
//         0xEF, 0xD3, 0x34, 0x59, 0x92, 0x40, 0xBB, 0x75, 0xF2, 0xE2,
//         0x99, 0xC5, 0x16, 0xD3, 0x6C, 0xBD, 0xC3, 0xFB, 0x22, 0xA6,
//         0x0E, 0x33, 0x67, 0x8E, 0xF8, 0x65, 0x5A, 0x51, 0x5D, 0xC4,
//         0xCD, 0x4D, 0xE4, 0xD7, 0x1D, 0x2D, 0x8A, 0x80, 0x03, 0x0F,
//         0x85, 0xB6, 0xFD, 0x14, 0xFD, 0xBB, 0xD7, 0x30, 0xD1, 0x2D,
//         0xEB, 0xAD, 0x79, 0x39, 0x3D, 0x0D, 0x5B, 0x42, 0xAD, 0xF2,
//         0x31, 0x89, 0xB9, 0xA2, 0x14, 0x36, 0xD9, 0xCF, 0xD0, 0x10,
//         0x9F, 0x2A, 0x7B, 0xA2, 0xAB, 0xA0, 0x8A, 0x3C, 0xB6, 0xC5,
//         0xA8, 0xD8, 0x95, 0xFD, 0xE1, 0x18, 0xB3, 0x1A, 0x3A, 0xDA,
//         0x53, 0xC1, 0x26, 0x3F, 0x49, 0xA3, 0x7B, 0x55, 0xE6, 0x04,
//         0xFF, 0x94, 0xE4, 0x5E, 0xA7, 0x25, 0x70, 0xBD, 0x4A, 0x7B,
//         0x31, 0xBE, 0x2A, 0x3B, 0xB8, 0x20, 0x7A, 0xDE, 0xDB, 0x3F,
//         0x40, 0xDD, 0xF8, 0x99, 0xE3, 0x8E, 0x11, 0x30, 0xC9, 0xA5,
//         0x65, 0x68, 0x57, 0x29, 0xC8, 0x2B, 0xEE, 0x92, 0xF8, 0x0E,
//         0x0C, 0x68, 0x2D, 0xD3, 0x9D, 0xE9, 0x78, 0x16, 0x1B, 0x02,
//         0xAA, 0xD6, 0xBE, 0x6E, 0x06, 0x4C, 0x20, 0xDD, 0xB3, 0xC5,
//         0x21, 0x66, 0xDD, 0x5C, 0xD4, 0x8D, 0xDE, 0x34, 0x07, 0xC6,
//         0xB5, 0x0B, 0x5F, 0x32, 0xCF, 0x6B, 0x6D, 0x1D, 0xCC, 0x7B,
//         0x15, 0x2A, 0x70, 0x7C, 0xCD, 0xB7, 0x11, 0x2D, 0xC7, 0xCD,
//         0x00, 0xC5, 0x55, 0x23, 0xDD, 0xBA, 0x60, 0x8A, 0x34, 0xC7,
//         0xDD, 0xEF, 0x0A, 0x27, 0x64, 0x2E, 0x55, 0x40, 0x83, 0x70,
//         0xCD, 0xBB, 0xF2, 0x13, 0x87, 0xA2, 0x23, 0x38, 0xC5, 0x8A,
//         0x79, 0xD2, 0x2A, 0xBF, 0x87, 0x09, 0x4A, 0xB5, 0x08, 0xAF,
//         0x48, 0x29, 0x16, 0xDE, 0xDC, 0xD2, 0xC5, 0x17, 0x34, 0x02,
//         0x25, 0xF6, 0x3E, 0xE2, 0x87, 0x3D, 0x2F, 0x5F, 0x5C, 0x61,
//         0x79, 0x70, 0xA3, 0x1A, 0x00, 0x10, 0xBC, 0xCC, 0x8A, 0xED,
//         0xB6, 0x0E, 0x39, 0x2D, 0x5A, 0x28, 0x7E, 0x89, 0xA9, 0x20,
//         0x58, 0x3A, 0x54, 0x4B, 0xDC, 0xD7, 0x08, 0xBC, 0x30, 0x0F,
//         0x51, 0xDF, 0xB4, 0xC9, 0xE3, 0x3E, 0x3B, 0xF1, 0x41, 0xCE,
//         0x4B, 0xD1, 0x51, 0x12, 0x41, 0x0D, 0xE4, 0xC9, 0x72, 0x6B,
//         0xE8, 0x97, 0xCE, 0xB5, 0x86, 0x92, 0x37, 0x2B, 0xA8, 0x34,
//         0xC1, 0x2E, 0x00, 0x5C, 0x33, 0x7A, 0x54, 0xC5, 0x72, 0xC7,
//         0xE0, 0xEA, 0x1A, 0x88, 0xBC, 0x9B, 0xD3, 0xF3, 0xC8, 0x59,
//         0x20, 0x89, 0x64, 0x6E, 0x83, 0xD6, 0xB3, 0xF8, 0x86, 0x4A,
//         0x50, 0x68, 0x1F, 0xA9, 0x42, 0x99, 0x8A, 0xDB, 0x3C, 0xC0,
//         0x86, 0xBB, 0x92, 0x6D, 0x9C, 0x44, 0x20, 0x64, 0x80, 0x3D,
//         0xEE, 0x9D,
//     ]);
//
//     println!("hash_2: {:?}", hash_2);
// }

#[no_mangle]
pub extern "C" fn generate_hash_2(threads: libc::uint32_t, sec: libc::uint8_t, pub_key_size: libc::size_t, pub_key: *const libc::uint8_t) -> *const libc::uint8_t {
    let mut modifier = [0u8; 16];
    thread_rng().fill_bytes(&mut modifier);
    let modifier = modifier;

    let done = Arc::new(RwLock::new(false));
    let modifier = Arc::new(modifier.to_vec());
    let pub_key = Arc::new(unsafe { std::slice::from_raw_parts(pub_key as *const u8, pub_key_size as usize) }.to_vec());
    let hash_2 = Arc::new(RwLock::new(vec![]));

    let handles = (0..threads).map(|i| {
        let done = done.clone();
        let modifier = modifier.clone();
        let pub_key = pub_key.clone();
        let hash_2 = hash_2.clone();

        thread::spawn(move || {
            if let Some(m) = work(&done, sec, i as u32, threads, &modifier, &pub_key) {
                let mut done = done.write().unwrap();
                let mut hash_2 = hash_2.write().unwrap();
                *done = true;
                *hash_2 = m.to_vec();
            }
        })
    }).collect::<Vec<_>>();

    for h in handles {
        h.join().unwrap();
    }

    let hash_2 = hash_2.read().unwrap().to_vec();

    hash_2.as_ptr()
}

fn work(done: &RwLock<bool>, sec: u8, offset: u32, increment: u32, modifier: &[u8], pub_key: &[u8]) -> Option<Vec<u8>> {
    let mut modifier_mut = modifier.to_vec();

    increment_bytes(&mut modifier_mut, offset);

    let mut tail = vec![0u8; 9];
    tail.append(&mut pub_key.to_vec());
    let tail = tail;

    let mut sha1 = sha1::Sha1::new();

    loop {
        if *done.read().unwrap() {
            return None;
        }

        sha1.reset();
        sha1.update(&modifier_mut);
        sha1.update(&tail);

        if let None = sha1.digest().iter().take(2 * sec as usize).find(|x| *x != &0) {
            break;
        } else {
            increment_bytes(&mut modifier_mut, increment);
        }
    }

    Some(sha1.digest().iter().take(14).map(|x| *x).collect())
}

fn increment_bytes(b: &mut [u8], mut amount: u32) -> u32 {
    let mut i = b.len() - 1;

    while amount > 0 {
        amount += b[i] as u32;
        b[i] = amount as u8;
        amount /= 256;

        if i == 0 {
            break;
        }

        i -= 1;
    }

    if amount > 0 {
        increment_bytes(b, amount)
    } else {
        amount
    }
}