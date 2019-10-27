clean: checkapp
	sudo rm -r ./apps/$(APP)

deploy: checkapp $(APP) move cleanup

status-app:
	@echo pulling from git...
	git clone https://github.com/qwhcr/status-app-web.git ./.temp

move:
	mv ./.temp/build ./apps/$(APP)

cleanup:
	sudo rm -r ./.temp

checkapp:
ifndef APP
$(error APP not specified)
endif

