# Recursive wildcard function
rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

OUT_DIR=dist/tbc

# Make everything. Keep this first so it's the default rule.
$(OUT_DIR): elemental_shaman $(OUT_DIR)/lib.wasm $(OUT_DIR)/sim_worker.js

elemental_shaman: $(OUT_DIR)/elemental_shaman/index.js $(OUT_DIR)/elemental_shaman/index.css $(OUT_DIR)/elemental_shaman/index.html detailed_results

detailed_results: $(OUT_DIR)/detailed_results/index.js $(OUT_DIR)/detailed_results/index.css $(OUT_DIR)/detailed_results/index.html

clean:
	rm -f ui/core/proto/*.ts
	rm -f sim/core/proto/*.pb.go
	rm -rf dist

# Host a local server, for dev testing
host: $(OUT_DIR)
	# Intentionally serve one level up, so the local site has 'tbc' as the first
	# directory just like github pages.
	npx http-server $(OUT_DIR)/..

ui/core/proto/proto.ts: proto/*.proto
	mkdir -p $(OUT_DIR)/protobuf-ts
	cp -r node_modules/@protobuf-ts/runtime/build/es2015/* $(OUT_DIR)/protobuf-ts
	sed -i -E "s/from '(.*)';/from '\1\.js';/g" $(OUT_DIR)/protobuf-ts/*
	sed -i -E "s/from \"(.*)\";/from '\1\.js';/g" $(OUT_DIR)/protobuf-ts/*
	npx protoc --ts_opt generate_dependencies --ts_out ui/core/proto --proto_path proto proto/api.proto

$(OUT_DIR)/core/tsconfig.tsbuildinfo: $(call rwildcard,ui/core,*.ts) ui/core/proto/proto.ts
	npx tsc -p ui/core
	sed -i 's/@protobuf-ts\/runtime/\/tbc\/protobuf-ts\/index/g' $(OUT_DIR)/core/proto/*.js
	sed -i -E "s/from \"(.*?)(\.js)?\";/from '\1\.js';/g" $(OUT_DIR)/core/proto/*.js

# Generic rule for building index.js for any class directory
$(OUT_DIR)/%/index.js: ui/%/index.ts ui/%/*.ts $(OUT_DIR)/core/tsconfig.tsbuildinfo
	npx tsc -p $(<D) 

# Generic rule for building index.css for any class directory
$(OUT_DIR)/%/index.css: ui/%/index.scss ui/%/*.scss $(call rwildcard,ui/core,*.scss)
	mkdir -p $(@D)
	npx sass $< $@

# Generic rule for building index.html for any class directory
$(OUT_DIR)/%/index.html: ui/index_template.html
	$(eval title := $(shell echo $(shell basename $(@D)) | sed -r 's/(^|_)([a-z])/\U \2/g' | cut -c 2-))
	echo $(title)
	mkdir -p $(@D)
	cat ui/index_template.html | sed 's/@@TITLE@@/TBC $(title) Simulator/g' > $@

.PHONY: wasm
wasm: $(OUT_DIR)/lib.wasm

$(OUT_DIR)/sim_worker.js: ui/worker/sim_worker.js
	cp ui/worker/sim_worker.js $(OUT_DIR)

# TODO: make different wasm generators per spec
# TODO: how to make this understand 
$(OUT_DIR)/lib.wasm: sim/wasm/* sim/core/proto/api.pb.go $(filter-out $(wildcard sim/core/items/*), $(call rwildcard,sim,*.go))
	GOOS=js GOARCH=wasm go build --tags=elemental_shaman -o ./$(OUT_DIR)/lib.wasm ./sim/wasm/

# Just builds the server binary
elesimweb: sim/core/proto/api.pb.go $(filter-out $(wildcard sim/core/items/*), $(call rwildcard,sim,*.go))
	go build --tags=elemental_shaman -o simweb ./sim/web/main.go

# Starts up a webserver hosting the $(OUT_DIR)/ and API endpoints.
elerunweb: sim/core/proto/api.pb.go
	go run --tags=elemental_shaman ./sim/web/main.go

sim/core/proto/api.pb.go: proto/*.proto
	protoc -I=./proto --go_out=./sim/core ./proto/*.proto

.PHONY: items
items: sim/core/items/all.go

sim/core/items/all.go: generate_items/*.go $(call rwildcard,sim/core/proto,*.go)
	go run generate_items/*.go -outDir=sim/core/items

test: $(OUT_DIR)/lib.wasm
	go test ./...
