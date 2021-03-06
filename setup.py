from setuptools import setup
from clogd import __VERSION__

setup(
    name='clog',
    version=__VERSION__,
    packages=['clogd'],
    package_data={
        '': ['static/*.*', 'views/*.*'],
    },
    install_requires=[
        'zpgdb==0.4.2',
        'Bottle==0.12.13',
        'waitress==1.1.0',
        'PyYAML==3.12',
    ],
)
