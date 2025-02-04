from setuptools import setup, find_packages # type: ignore

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="xyphos-client",
    version="0.1.0",
    author="Harsh VARDHAN GOSWAMI",
    author_email="harsh@xyphos.io",
    description="A client library for Xyphos - a secure multi-tenant key management system",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/theboringhumane/xyphos",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Security :: Cryptography",
    ],
    python_requires=">=3.8",
    install_requires=[
        "requests>=2.31.0",
        "urllib3>=2.0.0",
        "pydantic>=2.0.0",
        "cryptography>=42.0.0",
    ],
    extras_require={
        "dev": [
            "pytest>=8.0.0",
            "pytest-cov>=4.1.0",
            "pytest-asyncio>=0.23.0",
            "black>=24.1.0",
            "isort>=5.13.0",
            "mypy>=1.8.0",
            "ruff>=0.2.0",
        ],
    },
) 