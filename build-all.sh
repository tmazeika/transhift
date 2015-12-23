#!/usr/bin/env bash

# config
export GOPATH=/home/bionicrm/Code/GoCode/
bin_name="transhift"
build_dir="build"
package="github.com/bionicrm/transhift"
# ------

include=( README.md LICENSE )
os=( darwin freebsd linux netbsd openbsd windows )
arch=( 386 amd64 arm )
no_arm=( darwin openbsd windows )
no_tar=( darwin windows )

wd=$(pwd)

if [ ! -d ${build_dir} ]; then
    mkdir ${build_dir}
fi

cd ${build_dir}

for i in "${include[@]}"; do
    cp ${wd}/${i} ${i}
done

for o in "${os[@]}"; do
    for a in "${arch[@]}"; do
        if [[ "${no_arm[@]}" =~ ${o} ]] && [ ${a} == "arm" ]; then
            continue
        fi

        echo -n "Building for ${o}_${a}... "
        export GOOS=${o}
        export GOARCH=${a}
        /usr/local/go/bin/go build -o ${bin_name} ${package}

        if [ -f ${bin_name} ]; then
            if [[ "${no_tar[@]}" =~ ${o} ]]; then
                zip transhift_${o}_${a}.zip "${include[@]}" ${bin_name} > /dev/null
            else
                tar -zcf transhift_${o}_${a}.tar.gz "${include[@]}" ${bin_name}
            fi

            rm ${bin_name}
            echo "done"
        else
            echo "failed"
        fi
    done
done

for i in "${include[@]}"; do
    rm ${i}
done

cd ${wd}
