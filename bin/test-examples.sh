#/bin/bash

# Check the script is called with a directory as $1
(( $# )) || { echo "Usage: $0 <track-repo> [exercise-path ...]"; exit 1; }
track_repo="${1%/}"
[[ -d "${track_repo}" ]] || { echo "Usage: $0 <track-repo> [exercise-path ...]"; exit 1; }
shift

run_tests () {
    if (( $# != 0 )); then
        exercises="$@"
    else
        readarray -t exercises < <(
            jq -r --arg prefix "${track_repo}/exercises" '
                .exercises |
                (
                    (.practice | map(.slug = "practice/\(.slug)")) +
                    (.concept  | map(.slug = "concept/\(.slug)"))
                ) |
                map(
                    select(.status | IN("deprecated", "wip") | not) |
                    "\($prefix)/\(.slug)"
                ) |
                sort[]
            ' "${track_repo}/config.json"
        )
    fi
    for exercise in "${exercises[@]}"; do
        src="$(jq -r '.files|.example//.exemplar|.[0]' "$exercise/.meta/config.json")"
        dst="$(jq -r '.files.solution[0]' "$exercise/.meta/config.json")"
        cp "${exercise}/${src}" "${exercise}/${dst}"
        if ! output=$(go run . "$exercise" "${PWD}" 2>&1) || [[ -n "${output}" ]]; then
            printf 'Exercise %s does not test cleanly.\n%s\n' "${exercise}" "${output}"
        fi
        git -C "$exercise" checkout HEAD -- "${dst}"
    done
}

run_tests "$@"
