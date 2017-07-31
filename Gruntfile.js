module.exports = function (grunt) {
    grunt.initConfig({
        pkg: grunt.file.readJSON('package.json'),

        coffee: {
            compile: {
                files: {
                    'tmp/js/main.js': './js/main.coffee',
                }
            },
        },

        uglify: {
            main: {
                src: 'tmp/js/main.js',
                dest: 'js/main.js'
            },
        },

        watch: {
            js: {
                files: ['./js/*.coffee'],
                tasks: ['coffee', 'uglify'],
                options: {
                    spawn: false,
                }
            },
            html: {
                files: ['./view/*.html']
            },
            options: {
                livereload: true,
            },
        },
    });

    grunt.loadNpmTasks('grunt-contrib-watch');

    grunt.loadNpmTasks('grunt-contrib-coffee');
    grunt.loadNpmTasks('grunt-contrib-uglify');

    grunt.registerTask('default', ['coffee', 'uglify', 'watch']);

}
