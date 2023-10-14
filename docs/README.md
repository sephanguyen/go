# PlantUML
### setup docker for faster local render
Run this command every time you writing plantuml
```sh
docker run -d -p 8080:8080 plantuml/plantuml-server:jetty
```
### config vscode plantuml plugin
```
    "plantuml.render": "PlantUMLServer",
    "plantuml.server": "http://localhost:8080",
    "plantuml.urlFormat": "svg",
    "plantuml.exportFormat": "svg"
```
# CML
### VS Code
- For vscode, we can install the [Context Mapper extension](https://marketplace.visualstudio.com/items?itemName=contextmapper.context-mapper-vscode-extension)