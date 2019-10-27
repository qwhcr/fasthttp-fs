clean: checkapp
	sudo rm -r ./apps/$(APP)

deploy: checkapp $(APP) build move cleanup

status-app:
	@echo pulling from git...
	git clone https://github.com/qwhcr/status-app-web.git ./.temp

build: install-dep
	npm run build --prefix ./.temp

install-dep:
	npm i --prefix ./.temp

move:
	mv ./.temp/build ./apps/$(APP)

cleanup:
	sudo rm -r ./.temp

checkapp:
ifndef APP
$(error APP not specified)
endif

