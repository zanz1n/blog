#!/usr/bin/env bash

checksums=""

for filename in ${DIR}/*; do
    filechecksum=$(cd $DIR && sha256sum ${filename#"$DIR/"})
    checksums+=$filechecksum
    echo -e "$filechecksum" >"$filename.sha256"
    checksums+="\n"
done

echo -e "\n#### SHA256 Checksum\n\`\`\`\n$checksums\`\`\`" >>./RELEASE_CHANGELOG
echo -e "$checksums" >checksums.txt
