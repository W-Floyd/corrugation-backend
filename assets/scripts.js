document.addEventListener('alpine:init', () => {
    Alpine.store('isLoading', false)

    Alpine.store('newEntityDialog', {
        opened: false,
        entity: {},

        targetLocation: null,

        files: null,

        init() {
            this.reset()
        },

        reset() {
            this.opened = false
            this.files = null
            this.entity = {
                name: null,
                description: null,
                artifacts: null,
                location: null,
                metadata: {
                    quantity: null,
                    owners: null,
                    tags: null,
                },
            }
        },

        open(x) {
            this.targetLocation = x
            this.reset()
            document.getElementById("name").focus();
            this.opened = true
        },

        make() {
            if (this.files != null) {
                Alpine.store('api').uploadArtifacts()
            }
            this.entity.location = this.targetLocation
            this.entity.metadata.quantity = parseInt(this.entity.metadata.quantity)
            return this.entity
        }

    })


    Alpine.store('moveEntityDialog', {
        opened: false,

        sourceEntity: null,
        targetLocation: null,

        init() {
            this.reset()
        },

        reset() {
            this.opened = false
            this.targetLocation = null
        },

        open(x) {
            this.sourceEntity = x
            this.reset()
            this.targetLocation = Alpine.store('entities').fullstate.entities[x].location
            this.opened = true
        },

        move() {
            Alpine.store('api').moveEntity(this.sourceEntity, this.targetLocation)
        },

        getNotContains() {
            ret = []
            for (const key in Alpine.store('entities').fullstate.entities) {
                target = key
                shouldPush = true
                while (target != 0) {
                    if (target == this.sourceEntity) {
                        shouldPush = false
                    }
                    target = Alpine.store('entities').fullstate.entities[target].location
                }
                if (shouldPush) {
                    ret.push(key)
                }
            }
            return ret
        },

        formatOption(x) {
            tree = []
            target = x
            while (target != 0) {
                if (Alpine.store('entities').fullstate.entities[target].name == null || Alpine.store('entities').fullstate.entities[target].name == '') {
                    tree.push(target)
                } else {
                    tree.push(Alpine.store('entities').fullstate.entities[target].name)
                }

                target = Alpine.store('entities').fullstate.entities[target].location
            }
            tree.push('World')
            tree.reverse()
            return tree.join('/')
        },

    })

    Alpine.store('api', {
        newEntity(data) {
            // Creating a XHR object
            let xhr = new XMLHttpRequest();
            let url = "/api/entity";

            // open a connection
            xhr.open("POST", url, false);

            // Set the request header i.e. which type of content you are sending
            xhr.setRequestHeader("Content-Type", "application/json");

            // Create a state change callback
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 4 && xhr.status == 200) {
                    return xhr.status
                }
            };

            // Converting JSON data to string
            var data = JSON.stringify(Alpine.store('newEntityDialog').make());

            // Sending data with the request
            xhr.send(data);
        },

        moveEntity(x, y) {
            // Creating a XHR object
            let xhr = new XMLHttpRequest();
            let url = "/api/entity/" + x.toString();

            // open a connection
            xhr.open("PATCH", url, false);

            // Set the request header i.e. which type of content you are sending
            xhr.setRequestHeader("Content-Type", "application/json");

            // Create a state change callback
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 4 && xhr.status == 200) {
                    return xhr.status
                }
            };

            // Converting JSON data to string
            var data = JSON.stringify({ location: parseInt(y) });

            // Sending data with the request
            xhr.send(data);
        },

        delete(id) {
            // Creating a XHR object
            let xhr = new XMLHttpRequest();
            let url = "/api/entity/" + id;

            // open a connection
            xhr.open("DELETE", url, false);

            // Create a state change callback
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 4 && xhr.status == 200) {
                    return xhr.status
                }
            };

            xhr.send();
            Alpine.store('entities').load();
            Alpine.store('entities').loadLocationTree();
        },

        uploadArtifacts() {

            for (key in Alpine.store('newEntityDialog').files) {

                // Create FormData instance
                const fd = new FormData();
                fd.append('file', Alpine.store('newEntityDialog').files[key]);

                // Creating a XHR object
                let xhr = new XMLHttpRequest();
                let url = "/api/artifact";

                // Create a state change callback
                xhr.onreadystatechange = function () {
                    if (xhr.readyState === 4 && xhr.status == 200) {
                        let response = JSON.parse(xhr.responseText)
                        if (Alpine.store('newEntityDialog').entity.artifacts == null) {
                            Alpine.store('newEntityDialog').entity.artifacts = []
                        }
                        Alpine.store('newEntityDialog').entity.artifacts.push(parseInt(response))
                    }
                };

                // open a connection
                xhr.open("POST", url, false);

                // Sending data with the request
                xhr.send(fd);
            }

        }

    })

    Alpine.store('entities', {
        init() {
            this.currentEntity = 0
            this.load()
            this.loadLocationTree()
            this.loadFullState()
            Alpine.store('isLoading').this = false
        },

        currentEntity: 0,

        setCurrentEntity(x) {
            this.currentEntity = x
            this.load()
            this.loadLocationTree()
            this.loadFullState()
        },

        entities: {},

        load() {
            let url = '/api/entity/find/children/' + this.currentEntity + '/full';

            readAll(url)
                .then(response => response.json())
                .then(response => { this.entities = response; });
        },

        locationtree: [],
        fullentities: {},
        fullstate: {},

        loadFullState() {

            // Creating a XHR object
            let xhr = new XMLHttpRequest();
            let url = '/api/';

            // Create a state change callback
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 4 && xhr.status == 200) {
                    let response = JSON.parse(xhr.responseText)
                    Alpine.store('entities').fullstate = response
                }
            };

            // open a connection
            xhr.open("GET", url, false);

            // Sending data with the request
            xhr.send();

        },

        selectImages(x) { // Returns the IDs of the artifacts
            if (Alpine.store('entities').fullstate.entities[x].artifacts != null && Alpine.store('entities').fullstate.entities[x].artifacts.length > 0) {
                images = []
                for (key in Alpine.store('entities').fullstate.entities[x].artifacts) {
                    val = Alpine.store('entities').fullstate.entities[x].artifacts[key]
                    if (Alpine.store('entities').fullstate.artifacts[val].image == true) {
                        images.push(val)
                    }
                }
                return images
            }
            return null
        },

        readname(x) {
            if (x == 0) {
                return 'World'
            }
            return this.fullentities[x].name
        },

        loadLocationTree() {
            let url = '/api/entity';

            readAll(url)
                .then(response => response.json())
                .then(response => { this.fullentities = response; })
            this.locationtree = []
            this.recurseLocationTree(this.currentEntity)
            this.locationtree.reverse()
        },

        recurseLocationTree(x) {
            this.locationtree.push(x)
            if (x != 0) {
                elem = this.fullentities[x]
                this.recurseLocationTree(elem.location)
            }
        },

        hasChildren(x) {
            for (key in this.fullentities) {
                if (this.fullentities[key].location == x) {
                    return true
                }
            }
            return false
        },

        listChildLocations(x) {
            let childLocations = []
            for (key in this.fullentities) {
                if (this.fullentities[key].location == x) {
                    childLocations.push(this.fullentities[key])
                }
            }
            return childLocations
        }

    })

})

function dprint(x) {
    console.log(x)
}

function readAll(x) {
    return fetch(x)
}