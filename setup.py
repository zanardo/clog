from setuptools import setup
__version__ = '0.5beta1'

setup(
    name='clog',
    version=__version__,
    packages=['clogd'],
    package_data={
        '': ['static/*.*', 'views/*.*'],
    },
    install_requires=[
        'zpgdb==0.4',
        'Bottle==0.12.9',
        'waitress==1.0.0',
        'PyYAML==3.11',
    ],
)
